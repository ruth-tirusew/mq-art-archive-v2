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

Signed-in readers see **Apply as artist** in the site header. The application is a three-step form:

1. **About you** — artist name and a requested profile handle (3–30 lowercase letters, numbers or underscores). Availability is checked as you type.
2. **Your work** — Instagram, website, Telegram and/or a short note about your practice. At least one is required; these are only seen by the review team.
3. **Review** — confirm and submit.

Institution applications are not offered in the UI for now.

### 3. Admin review

An administrator approves or rejects the application from the admin app. A note to the applicant is required when rejecting and optional when approving. The applicant's own links are kept separately from the reviewer's note. Review criteria may include:

- authenticity of the account and the work
- validity of the requested handle
- compatibility with platform rules

The applicant gets an email with the decision (sent best-effort; logged only if `RESEND_API_KEY` is not set) and can follow progress on `/apply/status`.

### 4. Approval and going live

Approval is the only review. When it is approved:

- the user becomes an artist and a **private draft** profile is created with their name and handle
- the studio opens with a welcome guide: add a portrait and bio, optionally publish a first work, then tick **Make my profile public** and save

Publishing requires a bio and a portrait and needs no further approval. Unticking the box takes the profile offline.

If rejected, the applicant sees the reviewer's reason and can edit and apply again; the form is pre-filled.

## Required fields and validation rules

The onboarding form must validate all of the following:

- email or account identity must be present
- display name is required
- handle is required for artist/institution profile creation
- duplicate handle conflicts are rejected
- the user must not submit a malformed or empty profile request
- the application must be traceable to the authenticated user

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
