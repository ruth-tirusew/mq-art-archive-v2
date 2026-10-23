<script lang="ts">
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import { onboardingService } from "$lib/application/onboarding";
  import { ApiError } from "$lib/adapters/api/client";

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
    } catch {
      // No earlier application — start fresh.
    }
  });

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
  <p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">
    Artist application
  </p>
  <h1 class="mt-4 font-display text-4xl text-foreground">
    {reapplying ? "Apply again" : "Join Artiv as an artist"}
  </h1>
  <p class="mt-4 text-sm leading-relaxed text-muted-foreground">
    Artiv is a home for Ethiopian artists to share their work. Tell us a little
    about you, and our team will take a look — usually within a few days.
    You'll get an email when there's a decision, and once approved you can set
    up your public profile and publish right away.
  </p>

  <ol class="mt-8 flex items-center gap-3 font-mono text-[10px] uppercase tracking-[0.2em]">
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
    class="mt-10 space-y-5"
    onsubmit={(e) => {
      e.preventDefault();
      if (step < 3) step += 1;
      else void submit();
    }}
  >
    {#if step === 1}
      <div>
        <label class="label" for="display_name">Your artist name</label>
        <input
          id="display_name"
          class="field"
          bind:value={displayName}
          autocomplete="name"
          required
        />
        <p class="hint">The name shown on your profile. It can be a studio or stage name.</p>
      </div>
      <div>
        <label class="label" for="requested_handle">Choose your profile address</label>
        <div class="relative">
          <span
            class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-sm text-muted-foreground"
          >
            @
          </span>
          <input
            id="requested_handle"
            class="field handle-field {handleAvailable === false ||
            (normalizedHandle && !validHandle)
              ? 'border-destructive'
              : ''}"
            bind:value={requestedHandle}
            pattern={"[a-z0-9_]{3,30}"}
            minlength="3"
            maxlength="30"
            required
            autocomplete="off"
            aria-describedby="handle-status"
          />
        </div>
        <p
          id="handle-status"
          class="hint"
          class:text-destructive={handleAvailable === false ||
            (normalizedHandle && !validHandle)}
        >
          {#if normalizedHandle && !validHandle}
            Use 3–30 lowercase letters, numbers, or underscores.
          {:else if checkingHandle}
            Checking availability…
          {:else if handleAvailable === true}
            artiv/@{normalizedHandle} is available.
          {:else if handleAvailable === false}
            @{normalizedHandle} is already taken. Try adding your surname or a number.
          {:else if handleCheckFailed}
            We couldn't check availability right now. You can continue; we'll confirm it on submit.
          {:else}
            Your shareable link will look like artiv/@{normalizedHandle || "yourname"}. You can change it later in your studio.
          {/if}
        </p>
      </div>
    {:else if step === 2}
      <p class="text-sm text-muted-foreground">
        Help us see your work. Add at least one of the options below — a link to
        where your art lives is the quickest way to get approved.
      </p>
      <div>
        <label class="label" for="instagram">Instagram</label>
        <input id="instagram" class="field" placeholder="@yourhandle or link" bind:value={instagram} />
      </div>
      <div>
        <label class="label" for="website">Website or portfolio</label>
        <input id="website" class="field" placeholder="https://" bind:value={website} />
      </div>
      <div>
        <label class="label" for="telegram">Telegram</label>
        <input id="telegram" class="field" placeholder="@channel or link" bind:value={telegram} />
      </div>
      <div>
        <label class="label" for="about">Tell us about your practice</label>
        <textarea
          id="about"
          class="field min-h-28"
          placeholder="What do you make? How long have you been at it?"
          bind:value={about}
        ></textarea>
        <p class="hint">Only our review team sees this — it isn't published on your profile.</p>
      </div>
    {:else}
      <div class="rounded-sm border border-border/70 bg-card/30 p-6 text-sm">
        <p class="label">Review your application</p>
        <p class="mt-3 text-foreground">{displayName}</p>
        <p class="mt-1 font-mono text-xs text-accent">@{normalizedHandle}</p>
        <p class="mt-4 whitespace-pre-wrap text-muted-foreground">{notes}</p>
      </div>
      <p class="hint">
        By submitting you confirm this information is accurate. After review you'll
        get an email, and you can check progress any time on your application status page.
      </p>
    {/if}

    {#if error}
      <p class="text-sm text-destructive" role="alert">{error}</p>
    {/if}

    <div class="flex items-center gap-3">
      {#if step > 1}
        <button
          type="button"
          class="rounded-sm border border-border px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground transition hover:text-foreground"
          onclick={() => (step -= 1)}
        >
          Back
        </button>
      {/if}
      <button
        type="submit"
        class="rounded-sm border border-border bg-foreground px-5 py-3 font-mono text-[11px] uppercase tracking-[0.2em] text-background transition hover:opacity-90 disabled:opacity-50"
        disabled={submitting ||
          (step === 1 && !step1Ok) ||
          (step === 2 && !hasWork)}
      >
        {step < 3 ? "Continue" : submitting ? "Submitting…" : "Submit application"}
      </button>
    </div>
  </form>

  <p class="mt-8 text-sm text-muted-foreground">
    Already applied?
    <a href="/apply/status" class="text-accent underline underline-offset-4"
      >Check status</a
    >
  </p>
</section>

<style>
  .label {
    margin-bottom: 0.5rem;
    display: block;
    font-family: var(--font-mono, ui-monospace, monospace);
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.2em;
    color: var(--muted-foreground);
  }

  .hint {
    margin-top: 0.5rem;
    font-size: 0.75rem;
    color: var(--muted-foreground);
  }

  .field {
    width: 100%;
    border: 1px solid color-mix(in oklab, var(--border) 70%, transparent);
    background: color-mix(in oklab, var(--card) 30%, transparent);
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    color: var(--foreground);
  }

  .handle-field {
    padding-left: 2.5rem;
  }

  .field:focus {
    outline: 2px solid color-mix(in oklab, var(--accent) 50%, transparent);
    outline-offset: 2px;
  }
</style>
