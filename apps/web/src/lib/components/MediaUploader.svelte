<script lang="ts">
  import { uploadImage, validateImageFile } from '$lib/application/media';
  import type { UploadedMedia } from '$lib/core/domain/media';

  let {
    admin = false,
    onUploaded
  }: {
    admin?: boolean;
    onUploaded: (media: UploadedMedia) => void;
  } = $props();

  let uploading = $state(false);
  let progress = $state(0);
  let error = $state('');
  let previewUrl = $state('');
  let controller: AbortController | null = null;

  async function selectFile(event: Event) {
    const input = event.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;

    error = validateImageFile(file) ?? '';
    if (error) return;

    controller = new AbortController();
    uploading = true;
    progress = 0;
    try {
      const uploaded = await uploadImage(file, {
        admin,
        signal: controller.signal,
        onProgress: (value) => (progress = value)
      });
      previewUrl = uploaded.secure_url;
      onUploaded(uploaded);
    } catch (cause) {
      error =
        cause instanceof DOMException && cause.name === 'AbortError'
          ? 'Upload cancelled.'
          : cause instanceof Error
            ? cause.message
            : 'Image upload failed.';
    } finally {
      uploading = false;
      controller = null;
    }
  }
</script>

<div class="space-y-3">
  <label class="block">
    <span class="mb-2 block font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
      Upload image
    </span>
    <input
      type="file"
      accept="image/jpeg,image/png,image/webp"
      disabled={uploading}
      onchange={(event) => void selectFile(event)}
      class="block w-full text-sm text-muted-foreground file:mr-4 file:rounded-sm file:border file:border-border file:bg-transparent file:px-4 file:py-2 file:text-foreground disabled:opacity-50"
    />
  </label>

  {#if uploading}
    <div class="flex items-center gap-3">
      <progress class="h-2 flex-1 accent-current" max="100" value={progress}>{progress}%</progress>
      <span class="font-mono text-[10px] text-muted-foreground">{progress}%</span>
      <button
        type="button"
        class="font-mono text-[10px] uppercase tracking-[0.15em] text-destructive"
        onclick={() => controller?.abort()}
      >
        Cancel
      </button>
    </div>
  {/if}

  {#if error}
    <p class="text-sm text-destructive" role="alert">{error}</p>
  {/if}

  {#if previewUrl}
    <img src={previewUrl} alt="Uploaded preview" class="h-32 w-32 rounded-sm object-cover" />
  {/if}
</div>
