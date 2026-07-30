<script lang="ts">
	import { apiFetch } from '$lib/adapters/api/client';
	import { onMount } from 'svelte';

	type Role = 'public' | 'artist' | 'institution' | 'contributor' | 'admin';
	type User = { id: string; email: string; role: Role; display_name?: string };
	const roles: Role[] = ['public', 'artist', 'institution', 'contributor', 'admin'];
	let users = $state<User[]>([]);
	let error = $state('');

	onMount(async () => {
		try {
			users = await apiFetch<User[]>('/admin/v1/users?limit=100');
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Could not load users';
		}
	});

	async function changeRole(user: User, role: Role) {
		const updated = await apiFetch<User>(`/admin/v1/users/${user.id}`, {
			method: 'PATCH',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ role })
		});
		users = users.map((item) => (item.id === updated.id ? updated : item));
	}
</script>

<svelte:head><title>Users — Artiv Admin</title></svelte:head>

<section>
	<h1 class="text-2xl font-medium">Users</h1>
	<p class="mt-1 text-sm text-muted-foreground">Manage account roles and access.</p>
	{#if error}<p class="mt-6 text-destructive">{error}</p>{/if}
	<div class="mt-6 overflow-x-auto rounded-md border">
		<table class="w-full text-sm">
			<thead><tr class="border-b text-left"><th class="p-3">User</th><th class="p-3">Role</th></tr></thead>
			<tbody>
				{#each users as user}
					<tr class="border-b last:border-0">
						<td class="p-3"><p>{user.display_name || user.email}</p><p class="text-xs text-muted-foreground">{user.email}</p></td>
						<td class="p-3">
							<select class="rounded border bg-background px-2 py-1" value={user.role} onchange={(event) => changeRole(user, event.currentTarget.value as Role)}>
								{#each roles as role}<option value={role}>{role}</option>{/each}
							</select>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</section>
