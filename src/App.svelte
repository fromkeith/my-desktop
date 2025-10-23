<script lang="ts">
    import "./app.css";

    import { windowListProvider } from "$lib/pods/WindowsPod";
    import { authProvider } from "$lib/pods/AuthPod";

    import EmailListWindow from "$lib/my-components/EmailListWindow.svelte";
    import ComposeEmailWindow from "$lib/my-components/ComposeEmailWindow.svelte";
    import EmailContentsWindow from "$lib/my-components/EmailContentsWindow.svelte";
    import DesktopCommandBar from "$lib/my-components/DesktopCommandBar.svelte";

    import OAuth from "$lib/my-components/OAuth.svelte";

    const registery = {
        EmailListWindow: EmailListWindow,
        ComposeEmailWindow: ComposeEmailWindow,
        EmailContentsWindow: EmailContentsWindow,
    };

    $: auth = authProvider();
    $: authLoading = auth.isLoading;
    $: windows = windowListProvider();
</script>

<main class="w-screen h-screen">
    <DesktopCommandBar />
    {#each $windows as w (w.windowId)}
        <svelte:component this={registery[w.type]} window={w} {...w.props} />
    {/each}
    <!-- show login -->
    {#if $auth == null && !$authLoading}
        <div class="m-64">
            <OAuth />
        </div>
    {/if}

</main>

<style>
</style>
