# User onboarding guide

## Purpose

This document explains how a new user joins the platform, completes onboarding, and becomes eligible to publish content or manage a public profile.

## Who this applies to

This guide is for:

- Artists creating a public profile
- Institutions applying for a verified presence
- New users completing their first setup after account creation

## Onboarding flow

### 1. Create an account

1. Open the registration page.
2. Choose a sign-up method:
   - email and password
   - Google OAuth
3. Enter the required profile details.
4. Verify the email address if the system sends a verification link.
5. Sign in to continue.

### 2. Apply as an artist

Signed-in readers see **Apply as artist** in the site header. The application is a four-step form:

1. **About you**: artist name and a requested profile handle (3–30 lowercase letters, numbers or underscores). Availability is checked as you type.
2. **Your answers** (or **Your work**, see below).
3. **Your art**: add **a photo of yourself** and upload **at least five** pieces of your work, each with a title (medium and year are optional). Up to 20 are allowed. You choose and preview them on your device first; nothing is uploaded until you press **Submit** on the last step, when they go to Cloudinary (kept exactly as uploaded, shown to visitors as lossless WebP).
4. **Review**: confirm and submit.

**The portrait.** An artist application also needs a portrait (a clear photo of the applicant). It follows the same path as the works: chosen and previewed on the device, kept in the browser (IndexedDB) if the page is refreshed, uploaded only on **Submit**. It is not one of the five works and is never published as an artwork. It is frozen onto the application like the works, shown to the reviewer, and on approval becomes the profile photo, unless the artist already has one. The API refuses a submission without it (`400`), and it is locked while the application is under review.

**The five-works rule.** An artist application can't be submitted with fewer than five works; the form says how many more are needed and the API refuses the submission (`400`) otherwise. This applies to invited applicants too. The works are copied onto the application when it is submitted, so a reviewer always sees what was sent, even if the applicant later changes their works. While an application is under review the works are locked; after a rejection the applicant can edit them and apply again. Institution applications aren't affected.

**The questionnaire.** Once an admin has published a questionnaire revision (Admin → Applications → Questionnaire), step 2 is that questionnaire instead of the free-text links. Applicants answer in English, Amharic or both:

- Long answers use a small rich-text editor (bold, italic, links, lists, quotes). Everything is sanitised on the server before it is stored or shown.
- Answers **autosave** as the applicant types, and survive a reload or leaving the page. A bad answer (over the length limit, or a link that isn't `http(s)`) is flagged in place and never blocks the other answers from saving.
- Each question says whether the answer is **shown on the public profile** or **only seen by our team**.
- Required questions are enforced when the applicant continues. While an application is under review its answers are read-only; after a rejection they can be edited and resubmitted.

Before any revision is published, step 2 is the original form: Instagram, website, Telegram and/or a short note (at least one required), seen only by the review team.

Institution applications are not offered in the UI for now.

**Invitations.** For artists the team already knows, an admin can send an invite (Admin → Applications → Invites) with an optional name and a reserved handle. The link (`/apply?invite=…`) works once and expires (14 days by default).

The invitee **does not need an account**. The link opens a welcome page where they choose a password; that creates their account and signs them in. Because the link was sent only to that address, the account is created already verified, so there is no verification email to wait for. The address always comes from the invite, never from what the visitor types. If that address already has an account, the page says so and sends them to sign in instead; an invite never verifies or takes over an existing account. Google sign-in works too, as long as it is the invited address.

The form is then pre-filled with their name and handle. An invite does not skip review: the application is still approved by an admin. The link is shown to the admin once, and also emailed best-effort. A forwarded link can be used by whoever opens it first, so admins can revoke or resend a link at any time.

### 3. Admin review

An administrator approves or rejects the application from the admin app. A note to the applicant is required when rejecting and optional when approving. The applicant's own links are kept separately from the reviewer's note. Review criteria may include:

- authenticity of the account and the work
- validity of the requested handle
- compatibility with platform rules

The applicant gets an email with the decision (sent best-effort; logged only if `RESEND_API_KEY` is not set) and can follow progress on `/apply/status`.

### 4. Approval and going live

Reviewers see the applicant's answers (including those marked team-only) next to the approve and reject buttons. Approval is the only review. When it is approved:

- the user becomes an artist and their profile **goes live straight away** under the handle they asked for
- each of the (at least five) works they submitted becomes a published artwork on their page, with its blur hash
- the studio opens with a welcome guide: add a portrait and a short bio, and add more work when they like

Approval is the only gate: an admin has already reviewed the artist and their work, so there is no separate publish step. The profile goes live with the portrait the artist applied with; the bio is still the artist's to add (or shows their questionnaire answers until they do). An artist can still untick **Make my profile public** in the studio to take the profile offline, and admins can take one down too. Applications that predate the five-works rule go live with no artworks. If publishing fails part-way, the application goes back to pending and approving again finishes the job without duplicating anything.

If rejected, the applicant sees the reviewer's reason and can edit and apply again; the form is pre-filled.

### Questions after approval

Approved artists can edit their answers any time under **Studio → Questions**; edits appear on their public profile right away, and nothing is required any more. When an admin publishes a substantive revision (a question added, removed, or its rules changed), they may **nudge** approved artists: one optional email, plus a dismissible banner in the studio, inviting them to answer what is new. Nobody is emailed twice for the same revision, and ignoring a nudge changes nothing. Minor revisions (typos, wording) never nudge.

## Required fields and validation rules

The onboarding form must validate all of the following:

- email or account identity must be present
- display name is required
- handle is required for artist/institution profile creation
- duplicate handle conflicts are rejected
- the user must not submit a malformed or empty profile request
- the application must be traceable to the authenticated user
- an artist application needs at least five works, each with a title and an image the applicant uploaded themselves, and a portrait the applicant uploaded themselves

## User experience expectations

A successful onboarding experience should feel clear and low-friction:

- the user understands the next action at every step
- the system gives immediate feedback when a handle is unavailable
- every error message is specific and actionable
- successful approval is confirmed by a direct status update

## Success criteria

The onboarding process is considered complete when all of the following are true:

- the user can create and verify an account
- the user can submit a valid onboarding application
- the handle check prevents duplicates
- an admin can review and decide on the application
- approved users gain access to the intended profile or studio features

## Support and escalation

If a user cannot complete onboarding, support should check:

- whether the verification email was received and used
- whether the handle is already in use
- whether the application is pending or rejected
- whether the user is authenticated before submission

## Related documentation

- [docs/operations.md](operations.md)
- [README.md](../README.md)
