<script lang="ts">
  let { data } = $props();
</script>

<svelte:head>
  <title>Pending applications — mq admin</title>
</svelte:head>

<h1>Pending applications</h1>

{#if data.applications.length === 0}
  <p>No pending applications.</p>
{:else}
  <ul>
    {#each data.applications as application}
      <li>
        <a href="/applications/{application.id}">
          {application.display_name}
        </a>
        <span> ({application.applicant_type})</span>
      </li>
    {/each}
  </ul>
{/if}
