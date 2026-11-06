<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { authService, currentUser } from "$lib/application/auth";
  import { inviteService } from "$lib/application/invites";
  import { onboardingService } from "$lib/application/onboarding";
  import { questionnaireService } from "$lib/application/questionnaire";
  import { ApiError } from "$lib/adapters/api/client";
  import QuestionnaireForm from "$lib/components/questionnaire/QuestionnaireForm.svelte";
  import SafeHtml from "$lib/components/SafeHtml.svelte";
  import type { InvitePreview } from "$lib/core/domain/invite";
  import type { FormQuestion, QuestionnaireForm as QuestionnaireFormData } from "$lib/core/domain/questionnaire";
  import { answerText, formatFor, isBlankAnswer, linkLabel, safeHttpUrl } from "$lib/utils/answers";
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

  // The questionnaire replaces the free-text "Your work" step once an admin has published one.
  // "none" means no questionnaire is published yet (the plain form applies). "error" means we
  // couldn't find out, and applying on the plain form then would just fail at submit.
  let form = $state<QuestionnaireFormData | null>(null);
  let formState = $state<"loading" | "ready" | "none" | "error">("loading");
  let formRef = $state<ReturnType<typeof QuestionnaireForm>>();
  let reviewValues = $state<Record<string, string>>({});
  let savingStep = $state(false);

  // Invites arrive as /apply?invite=TOKEN.
  const inviteToken = $derived($page.url.searchParams.get("invite"));
  let invite = $state<InvitePreview | null>(null);
  let inviteMessage = $state("");
  let resendState = $state<"idle" | "sending" | "sent" | "failed">("idle");

  const inviteLive = $derived(invite?.status === "live");
  const emailMatches = $derived(
    !invite || !$currentUser || invite.email.toLowerCase() === $currentUser.email.toLowerCase(),
  );
  const needsVerify = $derived(inviteLive && !!$currentUser && !$currentUser.email_verified);
  // The invite is only claimed when the signed-in account is the invited one and is verified.
  const usingInvite = $derived(inviteLive && emailMatches && !needsVerify);

  // Fill gaps from the invite the first time it becomes usable. That can be well after load (for
  // example once the invitee verifies their email), so this is an effect rather than a one-off.
  // It never overwrites anything the applicant already has or has typed.
  let prefilledFromInvite = false;
  $effect(() => {
    if (!usingInvite || !invite || prefilledFromInvite) return;
    prefilledFromInvite = true;
    if (!displayName.trim()) displayName = invite.display_name ?? "";
    if (!requestedHandle.trim()) requestedHandle = invite.handle ?? "";
  });

  const stepLabels = $derived(["About you", formState === "none" ? "Your work" : "Your answers", "Review"]);

  const normalizedHandle = $derived(requestedHandle.trim().toLowerCase());
  const validHandle = $derived(/^[a-z0-9_]{3,30}$/.test(normalizedHandle));

  $effect(() => {
    const handle = normalizedHandle;
    clearTimeout(handleTimer);
    handleAvailable = null;
    handleCheckFailed = false;
    checkingHandle = false;
    if (!validHandle) return;
    // A live invite reserves its handle for the invitee, so the public check would call it taken.
    if (usingInvite && invite?.handle === handle) {
      handleAvailable = true;
      return;
    }

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

  const inviteProblems: Record<string, string> = {
    claimed: "This invitation has already been used. If that was you, check your application status.",
    expired: "This invitation has expired. Ask the Artiv team for a new link. You can still apply below.",
    revoked: "This invitation was cancelled. You can still apply below, or ask the team for a new link.",
  };

  async function loadInvite() {
    if (!inviteToken) return;
    try {
      invite = await inviteService.preview(inviteToken);
      if (invite.status !== "live") inviteMessage = inviteProblems[invite.status] ?? "";
    } catch (e) {
      invite = null;
      inviteMessage =
        e instanceof ApiError && e.status === 404
          ? "We couldn't find that invitation. Check the link, or apply below."
          : "We couldn't check your invitation just now. You can still apply below.";
    }
  }

  async function loadForm() {
    formState = "loading";
    try {
      form = await questionnaireService.getForm();
      formState = form ? "ready" : "none";
    } catch {
      form = null;
      formState = "error";
    }
  }

  async function resendVerification() {
    resendState = "sending";
    try {
      await authService.resendVerification();
      resendState = "sent";
    } catch {
      resendState = "failed";
    }
  }

  async function recheckVerification() {
    await authService.load();
  }

  // Pre-fill when re-applying after a rejection.
  onMount(async () => {
    void loadForm();
    void loadInvite();
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

  async function next() {
    attempted = true;
    if (step === 1 && !step1Ok) return;
    // Don't move past step 2 until we know which form applies.
    if (step === 2 && (formState === "loading" || formState === "error")) return;
    if (step === 2 && form) {
      // Show every problem, then make sure what's on screen is what we've saved.
      if (!formRef?.validate()) return;
      savingStep = true;
      const saved = await formRef.flush();
      savingStep = false;
      if (!saved) {
        error = "We couldn't save your answers just now. Check your connection and try again.";
        return;
      }
      error = "";
      reviewValues = formRef.current();
    } else if (step === 2 && !hasWork) {
      return;
    }
    attempted = false;
    step += 1;
  }

  const reviewQuestions = $derived(
    form ? [...form.revision.questions].sort((a, b) => a.position - b.position) : [],
  );

  function reviewText(q: FormQuestion): string {
    const body = reviewValues[q.key] ?? "";
    return q.type === "choice"
      ? answerText({ key: q.key, type: q.type, prompt_en: q.prompt_en, options: q.options, body, format: "plain", lang: "en" })
      : body;
  }

  async function submit() {
    submitting = true;
    error = "";
    try {
      if (usingInvite && inviteToken) {
        await inviteService.claim({
          token: inviteToken,
          display_name: displayName.trim(),
          requested_handle: normalizedHandle,
          // Same rule as below: with a questionnaire the answers are the application.
          notes: form ? "" : notes.trim(),
        });
      } else {
        await onboardingService.submit({
          applicant_type: applicantType,
          display_name: displayName.trim(),
          requested_handle: normalizedHandle,
          // With a questionnaire the answers are the application; the free-text notes only apply to the plain form.
          notes: form ? "" : notes.trim(),
        });
      }
      await goto("/apply/status");
    } catch (e) {
      if (e instanceof ApiError && e.status === 403 && usingInvite) {
        error = "This invitation was sent to a different email address, or your email isn't verified yet.";
      } else if (e instanceof ApiError && e.status === 409) {
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

  {#if invite && inviteLive}
    {#if !emailMatches}
      <Alert tone="warning" class="mt-6">
        This invitation was sent to <strong>{invite.email}</strong>, but you're signed in as
        <strong>{$currentUser?.email}</strong>. Sign in with the invited address to use it, or continue below with a regular application.
      </Alert>
    {:else if needsVerify}
      <Alert tone="warning" class="mt-6">
        <p>
          Please verify <strong>{$currentUser?.email}</strong> to use your invitation. We sent a link when you signed up.
        </p>
        <p class="mt-2 flex flex-wrap items-center gap-x-4 gap-y-1">
          <button type="button" class="underline underline-offset-4" onclick={resendVerification} disabled={resendState === "sending"}>
            {resendState === "sent" ? "Link sent. Check your inbox" : resendState === "failed" ? "Couldn't send. Try again" : "Send the link again"}
          </button>
          <button type="button" class="underline underline-offset-4" onclick={recheckVerification}>I've verified</button>
          <span class="text-muted-foreground">Or continue below as a regular application.</span>
        </p>
      </Alert>
    {:else}
      <Alert tone="success" class="mt-6">
        You're invited{invite.display_name ? `, ${invite.display_name}` : ""}. Answer a few questions and our team will review your application as usual.
      </Alert>
    {/if}
  {:else if inviteMessage}
    <Alert tone="warning" class="mt-6">{inviteMessage}</Alert>
  {/if}

  {#if priorCheckFailed}
    <Alert tone="warning" class="mt-6">
      We couldn't check for an earlier application. You can still apply; if you already have one, we'll tell you on submit.
    </Alert>
  {/if}

  <ol class="t-eyebrow mt-8 flex items-center gap-3 !text-[10px]" aria-label="Progress">
    {#each stepLabels as label, i}
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
      if (step < 3) void next();
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
    {:else if step === 2 && formState === "loading"}
      <p class="t-body text-muted-foreground" role="status">Loading the application form…</p>
    {:else if step === 2 && formState === "error"}
      <Alert tone="error">
        <p>We couldn't load the application form. Check your connection and try again.</p>
        <p class="mt-2">
          <button type="button" class="underline underline-offset-4" onclick={loadForm}>Try again</button>
        </p>
      </Alert>
    {:else if step === 2 && formState === "none"}
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
    {:else if step === 3}
      <div class="rounded-sm border border-border/70 bg-card/30 p-6 text-sm">
        <p class="t-eyebrow text-muted-foreground">Review your application</p>
        <p class="mt-3 text-foreground">{displayName}</p>
        <p class="t-meta mt-1 text-accent">@{normalizedHandle}</p>
        {#if form}
          <dl class="mt-6 space-y-5" data-testid="apply-review">
            {#each reviewQuestions as q (q.key)}
              {@const body = reviewValues[q.key] ?? ""}
              {#if !isBlankAnswer(body, formatFor(q))}
                <div>
                  <dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">{q.prompt_en}</dt>
                  <dd class="mt-1 text-foreground">
                    {#if formatFor(q) === "rich"}
                      <SafeHtml html={body} />
                    {:else if q.type === "link" && safeHttpUrl(body)}
                      <span class="break-all">{linkLabel(safeHttpUrl(body) ?? body)}</span>
                    {:else}
                      <span class="whitespace-pre-wrap">{reviewText(q)}</span>
                    {/if}
                  </dd>
                </div>
              {/if}
            {/each}
          </dl>
        {:else}
          <p class="mt-4 whitespace-pre-wrap text-muted-foreground">{notes}</p>
        {/if}
      </div>
      <p class="text-xs text-muted-foreground">
        By submitting you confirm this information is accurate. After review you'll get an email, and you can check progress any time on your application status page.
      </p>
    {/if}

    {#if form}
      <div class:hidden={step !== 2}>
        <QuestionnaireForm bind:this={formRef} {form} />
      </div>
    {/if}

    {#if error}<Alert tone="error">{error}</Alert>{/if}

    <div class="flex items-center gap-3">
      {#if step > 1}
        <Button variant="ghost" onclick={() => (step -= 1)}>Back</Button>
      {/if}
      <Button
        type="submit"
        loading={submitting || savingStep}
        disabled={step === 2 && (formState === "loading" || formState === "error")}
      >
        {step < 3 ? (savingStep ? "Saving…" : "Continue") : submitting ? "Submitting…" : "Submit application"}
      </Button>
    </div>
  </form>

  <p class="mt-8 text-sm text-muted-foreground">
    Already applied?
    <a href="/apply/status" class="text-accent underline underline-offset-4">Check status</a>
  </p>
</section>
