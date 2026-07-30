<script lang="ts">
  import { page } from '$app/stores';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { authService } from '$lib/application/auth';
  import { settingsService } from '$lib/application/settings';
  import type { ScrapeSettings } from '$lib/core/domain/settings';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { cn } from '$lib/utils';

  const inputClass =
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  const sectionIds = ['profile', 'account', 'notifications', 'integrations'] as const;
  type SectionId = (typeof sectionIds)[number];

  let { data } = $props();

  const activeSection = $derived.by((): SectionId => {
    const hash = $page.url.hash.replace(/^#/, '') as SectionId;
    return sectionIds.includes(hash) ? hash : 'profile';
  });

  // Profile
  let displayName = $state(data.user?.display_name ?? '');
  let avatarUrl = $state(data.user?.avatar_url ?? '');
  let profileSaving = $state(false);
  let profileMessage = $state('');
  let profileError = $state('');

  // Account
  let email = $state(data.user?.email ?? '');
  let emailPassword = $state('');
  let emailSaving = $state(false);
  let emailMessage = $state('');
  let emailError = $state('');

  let currentPassword = $state('');
  let newPassword = $state('');
  let confirmPassword = $state('');
  let passwordSaving = $state(false);
  let passwordMessage = $state('');
  let passwordError = $state('');

  // Notifications
  let emailOnNewApplication = $state(data.notifications.email_on_new_application);
  let emailOnEventSyncSummary = $state(data.notifications.email_on_event_sync_summary);
  let newsletterEnabled = $state(data.notifications.newsletter_enabled);
  let notifSaving = $state(false);
  let notifMessage = $state('');
  let notifError = $state('');

  // Integrations
  let scrape = $state<ScrapeSettings>(data.scrape);
  let scrapeEnabled = $state(data.scrape.scrape_enabled);
  let scrapeSources = $state(data.scrape.scrape_sources.join(', '));
  let scrapeUserAgent = $state(data.scrape.scrape_user_agent);
  let scrapeTimeout = $state(String(data.scrape.scrape_timeout_seconds));
  let scrapeInterval = $state(String(data.scrape.scrape_interval_seconds));
  let telegramEnabled = $state(data.scrape.telegram_enabled);
  let telegramApiId = $state(String(data.scrape.telegram_api_id || ''));
  let telegramApiHash = $state('');
  let telegramChannels = $state(data.scrape.telegram_channels.join(', '));
  let telegramKeywords = $state(data.scrape.telegram_keywords.join(', '));
  let telegramFetchLimit = $state(String(data.scrape.telegram_fetch_limit));
  let scrapeSaving = $state(false);
  let scrapeMessage = $state('');
  let scrapeError = $state('');

  function splitList(raw: string): string[] {
    return raw
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean);
  }

  function applyScrape(next: ScrapeSettings) {
    scrape = next;
    scrapeEnabled = next.scrape_enabled;
    scrapeSources = next.scrape_sources.join(', ');
    scrapeUserAgent = next.scrape_user_agent;
    scrapeTimeout = String(next.scrape_timeout_seconds);
    scrapeInterval = String(next.scrape_interval_seconds);
    telegramEnabled = next.telegram_enabled;
    telegramApiId = String(next.telegram_api_id || '');
    telegramApiHash = '';
    telegramChannels = next.telegram_channels.join(', ');
    telegramKeywords = next.telegram_keywords.join(', ');
    telegramFetchLimit = String(next.telegram_fetch_limit);
  }

  async function saveProfile(event: Event) {
    event.preventDefault();
    profileSaving = true;
    profileMessage = '';
    profileError = '';
    try {
      await authService.updateProfile({
        display_name: displayName.trim(),
        avatar_url: avatarUrl.trim()
      });
      profileMessage = 'Profile saved.';
    } catch (err) {
      profileError = err instanceof Error ? err.message : 'Save failed';
    } finally {
      profileSaving = false;
    }
  }

  async function saveEmail(event: Event) {
    event.preventDefault();
    emailSaving = true;
    emailMessage = '';
    emailError = '';
    try {
      const user = await authService.changeEmail(email.trim(), emailPassword);
      email = user.email;
      emailPassword = '';
      emailMessage = 'Email updated.';
    } catch (err) {
      emailError = err instanceof Error ? err.message : 'Update failed';
    } finally {
      emailSaving = false;
    }
  }

  async function savePassword(event: Event) {
    event.preventDefault();
    passwordSaving = true;
    passwordMessage = '';
    passwordError = '';
    if (newPassword !== confirmPassword) {
      passwordError = 'New passwords do not match.';
      passwordSaving = false;
      return;
    }
    try {
      await authService.changePassword(currentPassword, newPassword);
      currentPassword = '';
      newPassword = '';
      confirmPassword = '';
      passwordMessage = 'Password updated.';
    } catch (err) {
      passwordError = err instanceof Error ? err.message : 'Update failed';
    } finally {
      passwordSaving = false;
    }
  }

  async function saveNotifications(event: Event) {
    event.preventDefault();
    notifSaving = true;
    notifMessage = '';
    notifError = '';
    try {
      const prefs = await authService.updateNotifications({
        email_on_new_application: emailOnNewApplication,
        email_on_event_sync_summary: emailOnEventSyncSummary,
        newsletter_enabled: newsletterEnabled
      });
      emailOnNewApplication = prefs.email_on_new_application;
      emailOnEventSyncSummary = prefs.email_on_event_sync_summary;
      newsletterEnabled = prefs.newsletter_enabled;
      notifMessage = 'Notification preferences saved.';
    } catch (err) {
      notifError = err instanceof Error ? err.message : 'Save failed';
    } finally {
      notifSaving = false;
    }
  }

  async function saveScrape(event: Event) {
    event.preventDefault();
    scrapeSaving = true;
    scrapeMessage = '';
    scrapeError = '';
    try {
      const next = await settingsService.updateScrape({
        scrape_enabled: scrapeEnabled,
        scrape_sources: splitList(scrapeSources),
        scrape_user_agent: scrapeUserAgent.trim(),
        scrape_timeout_seconds: Number(scrapeTimeout) || 30,
        scrape_interval_seconds: Number(scrapeInterval) || 21600,
        telegram_enabled: telegramEnabled,
        telegram_api_id: Number(telegramApiId) || 0,
        telegram_api_hash: telegramApiHash.trim() || undefined,
        telegram_channels: splitList(telegramChannels),
        telegram_keywords: splitList(telegramKeywords),
        telegram_fetch_limit: Number(telegramFetchLimit) || 50
      });
      applyScrape(next);
      scrapeMessage = 'Integrations saved. Use Events → Sync now to fetch with the new config.';
    } catch (err) {
      scrapeError = err instanceof Error ? err.message : 'Save failed';
    } finally {
      scrapeSaving = false;
    }
  }
</script>

<svelte:head>
  <title>Settings — mq admin</title>
</svelte:head>

<PageHeader
  eyebrow="08 — Settings"
  title="Settings"
  description="Manage your admin profile, account security, notifications, and scrape integrations."
/>

{#if activeSection === 'profile'}
  <Card>
    <CardHeader>
      <CardTitle>Profile</CardTitle>
    </CardHeader>
    <CardContent>
      <form class="space-y-4" onsubmit={saveProfile}>
        {#if profileError}
          <Alert variant="destructive">{profileError}</Alert>
        {/if}
        {#if profileMessage}
          <Alert>{profileMessage}</Alert>
        {/if}
        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-2">
            <Label for="display-name">Display name</Label>
            <input id="display-name" class={inputClass} bind:value={displayName} />
          </div>
          <div class="space-y-2">
            <Label for="avatar-url">Avatar URL</Label>
            <input id="avatar-url" class={inputClass} bind:value={avatarUrl} />
          </div>
        </div>
        <div class="flex items-center gap-3">
          <span class="text-sm text-muted-foreground">Role</span>
          <Badge variant="secondary">{data.user?.role ?? 'admin'}</Badge>
        </div>
        <Button type="submit" disabled={profileSaving}>
          {profileSaving ? 'Saving…' : 'Save profile'}
        </Button>
      </form>
    </CardContent>
  </Card>
{:else if activeSection === 'account'}
  <div class="space-y-6">
    <Card>
      <CardHeader>
        <CardTitle>Change email</CardTitle>
      </CardHeader>
      <CardContent>
        {#if data.user && !data.user.has_password}
          <Alert>
            This account has no password (OAuth-only). Use forgot-password on the login page to set
            one before changing email.
          </Alert>
        {:else}
          <form class="space-y-4" onsubmit={saveEmail}>
            {#if emailError}
              <Alert variant="destructive">{emailError}</Alert>
            {/if}
            {#if emailMessage}
              <Alert>{emailMessage}</Alert>
            {/if}
            <div class="space-y-2">
              <Label for="email">Email</Label>
              <input id="email" type="email" class={inputClass} bind:value={email} required />
            </div>
            <div class="space-y-2">
              <Label for="email-password">Current password</Label>
              <input
                id="email-password"
                type="password"
                class={inputClass}
                bind:value={emailPassword}
                required
                autocomplete="current-password"
              />
            </div>
            <Button type="submit" disabled={emailSaving}>
              {emailSaving ? 'Updating…' : 'Update email'}
            </Button>
          </form>
        {/if}
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>Change password</CardTitle>
      </CardHeader>
      <CardContent>
        {#if data.user && !data.user.has_password}
          <Alert>
            No password is set for this account. Use forgot-password on the login page to create
            one.
          </Alert>
        {:else}
          <form class="space-y-4" onsubmit={savePassword}>
            {#if passwordError}
              <Alert variant="destructive">{passwordError}</Alert>
            {/if}
            {#if passwordMessage}
              <Alert>{passwordMessage}</Alert>
            {/if}
            <div class="space-y-2">
              <Label for="current-password">Current password</Label>
              <input
                id="current-password"
                type="password"
                class={inputClass}
                bind:value={currentPassword}
                required
                autocomplete="current-password"
              />
            </div>
            <div class="grid gap-4 md:grid-cols-2">
              <div class="space-y-2">
                <Label for="new-password">New password</Label>
                <input
                  id="new-password"
                  type="password"
                  class={inputClass}
                  bind:value={newPassword}
                  required
                  minlength={8}
                  autocomplete="new-password"
                />
              </div>
              <div class="space-y-2">
                <Label for="confirm-password">Confirm new password</Label>
                <input
                  id="confirm-password"
                  type="password"
                  class={inputClass}
                  bind:value={confirmPassword}
                  required
                  minlength={8}
                  autocomplete="new-password"
                />
              </div>
            </div>
            <Button type="submit" disabled={passwordSaving}>
              {passwordSaving ? 'Updating…' : 'Update password'}
            </Button>
          </form>
        {/if}
      </CardContent>
    </Card>
  </div>
{:else if activeSection === 'notifications'}
  <Card>
    <CardHeader>
      <CardTitle>Email &amp; notifications</CardTitle>
    </CardHeader>
    <CardContent>
      <form class="space-y-4" onsubmit={saveNotifications}>
        {#if notifError}
          <Alert variant="destructive">{notifError}</Alert>
        {/if}
        {#if notifMessage}
          <Alert>{notifMessage}</Alert>
        {/if}
        <label class="flex items-start gap-3 text-sm">
          <input type="checkbox" class="mt-1" bind:checked={emailOnNewApplication} />
          <span>
            <span class="font-medium text-foreground">New applications</span>
            <span class="mt-0.5 block text-muted-foreground">
              Email when an artist or institution submits an onboarding application.
            </span>
          </span>
        </label>
        <label class="flex items-start gap-3 text-sm">
          <input type="checkbox" class="mt-1" bind:checked={emailOnEventSyncSummary} />
          <span>
            <span class="font-medium text-foreground">Event sync summary</span>
            <span class="mt-0.5 block text-muted-foreground">
              Email a summary after scraper sync runs.
            </span>
          </span>
        </label>
        <label class="flex items-start gap-3 text-sm">
          <input type="checkbox" class="mt-1" bind:checked={newsletterEnabled} />
          <span>
            <span class="font-medium text-foreground">Platform newsletter</span>
            <span class="mt-0.5 block text-muted-foreground">
              Occasional product and curation updates.
            </span>
          </span>
        </label>
        <Button type="submit" disabled={notifSaving}>
          {notifSaving ? 'Saving…' : 'Save preferences'}
        </Button>
      </form>
    </CardContent>
  </Card>
{:else}
  <form class="space-y-6" onsubmit={saveScrape}>
    {#if scrapeError}
      <Alert variant="destructive">{scrapeError}</Alert>
    {/if}
    {#if scrapeMessage}
      <Alert>{scrapeMessage}</Alert>
    {/if}

    <Card>
      <CardHeader>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <CardTitle>Telegram MTProto</CardTitle>
          <Badge variant={scrape.session_authorized ? 'default' : 'secondary'}>
            {scrape.session_authorized ? 'Session authorized' : 'Session missing'}
          </Badge>
        </div>
      </CardHeader>
      <CardContent class="space-y-4">
        <p class="text-sm text-muted-foreground">
          Authorize the Telegram session with the <code class="font-mono text-xs">telegram-login</code>
          CLI. API hash is write-only — leave blank to keep the current value.
        </p>
        <label class="flex items-center gap-3 text-sm">
          <input type="checkbox" bind:checked={telegramEnabled} />
          Enable Telegram scraping
        </label>
        <div class="grid gap-4 md:grid-cols-2">
          <div class="space-y-2">
            <Label for="api-id">API ID</Label>
            <input id="api-id" class={inputClass} bind:value={telegramApiId} inputmode="numeric" />
          </div>
          <div class="space-y-2">
            <Label for="api-hash">API hash</Label>
            <input
              id="api-hash"
              type="password"
              class={inputClass}
              bind:value={telegramApiHash}
              placeholder={scrape.telegram_api_hash_set ? '•••••••• (unchanged if blank)' : ''}
              autocomplete="off"
            />
          </div>
          <div class="space-y-2 md:col-span-2">
            <Label for="channels">Channels (comma-separated)</Label>
            <input id="channels" class={inputClass} bind:value={telegramChannels} />
          </div>
          <div class="space-y-2 md:col-span-2">
            <Label for="keywords">Keywords (comma-separated, empty = all)</Label>
            <input id="keywords" class={inputClass} bind:value={telegramKeywords} />
          </div>
          <div class="space-y-2">
            <Label for="fetch-limit">Fetch limit</Label>
            <input id="fetch-limit" class={inputClass} bind:value={telegramFetchLimit} inputmode="numeric" />
          </div>
        </div>
      </CardContent>
    </Card>

    <Card>
      <CardHeader>
        <CardTitle>RSS scrape</CardTitle>
      </CardHeader>
      <CardContent class="space-y-4">
        <label class="flex items-center gap-3 text-sm">
          <input type="checkbox" bind:checked={scrapeEnabled} />
          Enable RSS scraping
        </label>
        <div class="space-y-2">
          <Label for="sources">Feed URLs (comma-separated)</Label>
          <input id="sources" class={inputClass} bind:value={scrapeSources} />
        </div>
        <div class="grid gap-4 md:grid-cols-3">
          <div class="space-y-2 md:col-span-1">
            <Label for="ua">User agent</Label>
            <input id="ua" class={inputClass} bind:value={scrapeUserAgent} />
          </div>
          <div class="space-y-2">
            <Label for="timeout">Timeout (seconds)</Label>
            <input id="timeout" class={inputClass} bind:value={scrapeTimeout} inputmode="numeric" />
          </div>
          <div class="space-y-2">
            <Label for="interval">Interval (seconds)</Label>
            <input id="interval" class={inputClass} bind:value={scrapeInterval} inputmode="numeric" />
          </div>
        </div>
      </CardContent>
    </Card>

    <div class="flex flex-wrap items-center gap-3">
      <Button type="submit" disabled={scrapeSaving}>
        {scrapeSaving ? 'Saving…' : 'Save integrations'}
      </Button>
      <a
        href="/events"
        class={cn('text-sm text-accent underline underline-offset-4 hover:text-foreground')}
      >
        Open Events to sync
      </a>
    </div>
  </form>
{/if}
