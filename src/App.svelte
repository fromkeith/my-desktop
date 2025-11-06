<script lang="ts">
    import "./app.css";

    import { windowListProvider } from "$lib/pods/WindowsPod";
    import { isAuthValidProvider } from "$lib/pods/AuthPod";

    import EmailListWindow from "$lib/my-components/EmailListWindow.svelte";
    import ComposeEmailWindow from "$lib/my-components/ComposeEmailWindow.svelte";
    import EmailContentsWindow from "$lib/my-components/EmailContentsWindow.svelte";
    import DesktopCommandBar from "$lib/my-components/DesktopCommandBar.svelte";
    import DesktopIcons from "$lib/my-components/DesktopIcons.svelte";

    import Progress from "$lib/components/ui/progress/progress.svelte";
    import LoginWindow from "$lib/my-components/LoginWindow.svelte";

    const registery: Record<string, ConstructorOfATypedSvelteComponent> = {
        EmailListWindow: EmailListWindow,
        ComposeEmailWindow: ComposeEmailWindow,
        EmailContentsWindow: EmailContentsWindow,
    };
    $: windows = windowListProvider();
    $: isAuthValid = isAuthValidProvider();
    $: authLoading = isAuthValid.isLoading;
    //
</script>

<main class="w-screen h-screen">
    .
    {#if !$isAuthValid}
        {#if $authLoading}
            <Progress value={null} />
        {:else}
            <LoginWindow />
        {/if}
    {:else}
        <DesktopCommandBar />
        <DesktopIcons />
        {#each $windows as w (w.windowId)}
            <svelte:component
                this={registery[w.type]}
                window={w}
                {...w.props}
            />
        {/each}
    {/if}
</main>

<style>
</style>
