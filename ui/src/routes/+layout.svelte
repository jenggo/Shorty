<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { auth } from '$lib/stores/auth';
  import { goto } from '$app/navigation';

  let { children } = $props();

  onMount(() => {
    auth.initialize();
  });

  $effect(() => {
    // Redirect to login if not authenticated
    if (!$auth.isAuthenticated && window.location.pathname !== '/login') {
      goto('/login');
    }
  });
</script>

{@render children()}
