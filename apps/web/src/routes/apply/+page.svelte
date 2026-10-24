<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { onboardingService } from "$lib/application/onboarding";
  import { ApiError } from "$lib/adapters/api/client";
  import { Alert, Button, Field, Input, PageHeader, Textarea } from "$ui";

  // Institutions are not onboarded through this form for now.
  const applicantType = "artist" as const;

  let step = $state(1);
  let instagram = $state("");
  let website = $state("");
  let telegram = $state("");
  let about = $state("");
  let displayName = $state("");
  let requestedHandle = $state("");
  let handleAvailable = $state<boolean | null>(null);
  let checkingHandle = $state(false);
  let handleCheckFailed = $state(false);
  let submitting = $state(false);
  let error = $state("");
  let reapplying = $state(false);
  let priorCheckFailed = $state(false);
  let attempted = $state(false);
  let handleTimer: ReturnType<typeof setTimeout> | undefined;

  const normalizedHandle = $derived(requestedHandle.trim().toLowerCase());
  const validHandle = $derived(/^[a-z0-9_]{3,30}$/.test(normalizedHandle));

  $effect(() => {
    const handle = normalizedHandle;
    clearTimeout(handleTimer);
    handleAvailable = null;
    handleCheckFailed = false;
    checkingHandle = false;
    if (!validHandle) return;

    handleTimer = setTimeout(async () => {
      checkingHandle = true;
      try {
        const result = await onboardingService.checkHandleAvailability(handle);
        if (normalizedHandle === handle) handleAvailable = result.available;
      } catch {
        if (normalizedHandle === handle) {
          handleAvailable = null;
          handleCheckFailed = true;
        }
      } finally {
        if (normalizedHandle === handle) checkingHandle = false;
      }
    }, 300);
    return () => clearTimeout(handleTimer);
  });

  const hasWork = $derived(
    Boolean(instagram.trim() || website.trim() || telegram.trim() || about.trim()),
  );
  const step1Ok = $derived(
    Boolean(displayName.trim()) &&
      validHandle &&
      (handleAvailable === true || handleCheckFailed),
  );

  const notes = $derived(
    [
      instagram.trim() && `Instagram: ${instagram.trim()}`,
      website.trim() && `Website: ${website.trim()}`,
      telegram.trim() && `Telegram: ${telegram.trim()}`,
      about.trim() && `About my practice:\n${about.trim()}`,
    ]
      .filter(Boolean)
      .join("\n"),
  );

  // Pre-fill when re-applying after a rejection.
  onMount(async () => {
    try {
      const previous = await onboardingService.getMyApplication();
      if (previous.status === "pending") {
        await goto("/apply/status");
        return;
      }
      if (previous.status === "rejected") {
        displayName = previous.display_name;
        requestedHandle = previous.requested_handle ?? "";
        about = previous.notes ?? "";
        reapplying = true;
      }
    } catch (e) {
      // 404 means no earlier application: start fresh. Anything else we can still recover from,
      // but say so instead of silently pretending there is no history.
      if (!(e instanceof ApiError && e.status === 404)) priorCheckFailed = true;
    }
  });

  const nameError = $derived(attempted && !displayName.trim() ? "Tell us the name to show on your profile." : undefined);
  const handleError = $derived.by(() => {
    if (normalizedHandle && !validHandle) return "Use 3–30 lowercase letters, numbers, or underscores.";
    if (handleAvailable === false) return `@${normalizedHandle} is already taken. Try adding your surname or a number.`;
    if (attempted && !normalizedHandle) return "Choose a profile address.";
    return undefined;
  });
  const handleHint = $derived.by(() => {
    if (checkingHandle) return "Checking availability…";
    if (handleAvailable === true) return `artiv/@${normalizedHandle} is available.`;
    if (handleCheckFailed) return "We couldn't check availability right now. You can continue; we'll confirm it on submit.";
    return `Your shareable link will look like artiv/@${normalizedHandle || "yourname"}. You can change it later in your studio.`;
  });
  const workError = $derived(attempted && !hasWork ? "Add at least one link, or a few lines about your work." : undefined);

  function next() {
    attempted = true;
    if (step === 1 && !step1Ok) return;
    if (step === 2 && !hasWork) return;
    attempted = false;
    step += 1;
  }

  async function submit() {
    submitting = true;
    error = "";
    try {
      await onboardingService.submit({
        applicant_type: applicantType,
        display_name: displayName.trim(),
        requested_handle: normalizedHandle,
        notes: notes.trim(),
      });
      await goto("/apply/status");
    } catch (e) {
      if (e instanceof ApiError && e.status === 409) {
        error =
          "We couldn't submit this. Either the handle was just taken or you already have an application under review. Check your application status, or try a different handle.";
      } else {
        error = e instanceof Error ? e.message : "Could not submit application";
      }
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>Apply — Artiv</title>
</svelte:head>

<section class="mx-auto max-w-2xl px-6 py-16 md:px-10 md:py-24">
  <PageHeader
    eyebrow="Artist application"
    lead="Artiv is a home for Ethiopian artists to share their work. Tell us a little about you, and our team will take a look — usually within a few days. You'll get an email when there's a decision, and once approved you can set up your public profile and publish right away."
  >
    {#if reapplying}Apply <em class="italic">again</em>{:else}Join Artiv as an <em class="italic">artist</em>{/if}
  </PageHeader>

  {#if priorCheckFailed}
    <Alert tone="warning" class="mt-6">
      We couldn't check for an earlier application. You can still apply; if you already have one, we'll tell you on submit.
    </Alert>
  {/if}

  <ol class="t-eyebrow mt-8 flex items-center gap-3 !text-[10px]" aria-label="Progress">
    {#each ["About you", "Your work", "Review"] as label, i}
      <li class="flex items-center gap-3">
        <span
          class="flex h-6 w-6 items-center justify-center rounded-full border {step === i + 1
            ? 'border-foreground bg-foreground text-background'
            : step > i + 1
              ? 'border-accent text-accent'
              : 'border-border text-muted-foreground'}"
          aria-current={step === i + 1 ? "step" : undefined}
        >
          {step > i + 1 ? "✓" : i + 1}
        </span>
        <span class={step === i + 1 ? "text-foreground" : "text-muted-foreground"}>{label}</span>
        {#if i < 2}<span class="h-px w-6 bg-border" aria-hidden="true"></span>{/if}
      </li>
    {/each}
  </ol>

  <form
    class="mt-10 space-y-6"
    novalidate
    onsubmit={(e) => {
      e.preventDefault();
      if (step < 3) next();
      else void submit();
    }}
  >
    {#if step === 1}
      <Field label="Your artist name" required error={nameError} hint="The name shown on your profile. It can be a studio or stage name.">
        {#snippet children({ id, describedby, invalid, required })}
          <Input {id} {required} {invalid} aria-describedby={describedby} bind:value={displayName} autocomplete="name" />
        {/snippet}
      </Field>
      <Field label="Choose your profile address" required error={handleError} hint={handleHint}>
        {#snippet children({ id, describedby, invalid, required })}
          <div class="relative">
            <span class="pointer-events-none absolute left-3.5 top-1/2 -translate-y-1/2 text-sm text-muted-foreground" aria-hidden="true">@</span>
            <Input
              {id}
              {required}
              {invalid}
              aria-describedby={describedby}
              class="pl-8"
              bind:value={requestedHandle}
              maxlength={30}
              autocomplete="off"
              autocapitalize="none"
              spellcheck={false}
            />
          </div>
        {/snippet}
      </Field>
    {:else if step === 2}
      <p class="t-body text-muted-foreground">
        Help us see your work. Add at least one of the options below — a link to where your art lives is the quickest way to get approved.
      </p>
      <Field label="Instagram">
        {#snippet children({ id })}<Input {id} placeholder="@yourhandle or link" bind:value={instagram} />{/snippet}
      </Field>
      <Field label="Website or portfolio">
        {#snippet children({ id })}<Input {id} type="url" placeholder="https://" bind:value={website} />{/snippet}
      </Field>
      <Field label="Telegram">
        {#snippet children({ id })}<Input {id} placeholder="@channel or link" bind:value={telegram} />{/snippet}
      </Field>
      <Field label="Tell us about your practice" error={workError} hint="Only our review team sees this — it isn't published on your profile.">
        {#snippet children({ id, describedby, invalid })}
          <Textarea {id} {invalid} aria-describedby={describedby} rows={5} placeholder="What do you make? How long have you been at it?" bind:value={about} />
        {/snippet}
      </Field>
    {:else}
      <div class="rounded-sm border border-border/70 bg-card/30 p-6 text-sm">
        <p class="t-eyebrow text-muted-foreground">Review your application</p>
        <p class="mt-3 text-foreground">{displayName}</p>
        <p class="t-meta mt-1 text-accent">@{normalizedHandle}</p>
        <p class="mt-4 whitespace-pre-wrap text-muted-foreground">{notes}</p>
      </div>
      <p class="text-xs text-muted-foreground">
        By submitting you confirm this information is accurate. After review you'll get an email, and you can check progress any time on your application status page.
      </p>
    {/if}

    {#if error}<Alert tone="error">{error}</Alert>{/if}

    <div class="flex items-center gap-3">
      {#if step > 1}
        <Button variant="ghost" onclick={() => (step -= 1)}>Back</Button>
      {/if}
      <Button type="submit" loading={submitting}>
        {step < 3 ? "Continue" : submitting ? "Submitting…" : "Submit application"}
      </Button>
    </div>
  </form>

  <p class="mt-8 text-sm text-muted-foreground">
    Already applied?
    <a href="/apply/status" class="text-accent underline underline-offset-4">Check status</a>
  </p>
</section>
