<script lang="ts">
    import "./app.css";

    import { windowListProvider } from "$lib/pods/WindowsPod";

    import EmailListWindow from "$lib/my-components/EmailListWindow.svelte";
    import ComposeEmailWindow from "$lib/my-components/ComposeEmailWindow.svelte";
    import EmailContentsWindow from "$lib/my-components/EmailContentsWindow.svelte";
    import DesktopCommandBar from "$lib/my-components/DesktopCommandBar.svelte";

    const registery = {
        EmailListWindow: EmailListWindow,
        ComposeEmailWindow: ComposeEmailWindow,
        EmailContentsWindow: EmailContentsWindow,
    };

    $: windows = windowListProvider();
</script>

<main class="w-screen h-screen">
    <DesktopCommandBar />
    {#each $windows as w (w.windowId)}
        <svelte:component this={registery[w.type]} window={w} {...w.props} />
    {/each}
</main>

<style>
</style>
