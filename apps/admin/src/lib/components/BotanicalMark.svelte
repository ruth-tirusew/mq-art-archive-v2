<script lang="ts">
  let { class: className = '' }: { class?: string } = $props();
</script>

<svg
  class={className}
  viewBox="0 0 120 140"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  aria-hidden="true"
>
  <g stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round">
    <path d="M60 128 V72" />
    <path d="M60 100 C48 92 38 78 36 62" />
    <path d="M60 92 C72 84 82 70 84 54" />
    <ellipse cx="60" cy="48" rx="10" ry="14" />
    <ellipse cx="44" cy="56" rx="9" ry="12" transform="rotate(-28 44 56)" />
    <ellipse cx="76" cy="56" rx="9" ry="12" transform="rotate(28 76 56)" />
    <ellipse cx="38" cy="72" rx="8" ry="11" transform="rotate(-48 38 72)" />
    <ellipse cx="82" cy="72" rx="8" ry="11" transform="rotate(48 82 72)" />
    <circle cx="60" cy="48" r="3.5" fill="currentColor" stroke="none" opacity="0.35" />
  </g>
</svg>
