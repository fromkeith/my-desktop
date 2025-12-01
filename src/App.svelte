<script lang="ts">
    import "./app.css";

    import { windowListProvider } from "$lib/pods/WindowsPod";
    import { isAuthValidProvider } from "$lib/pods/AuthPod";
    import { WindowType } from "$lib/models";

    import EmailListWindow from "$lib/my-components/EmailListWindow.svelte";
    import ComposeEmailWindow from "$lib/my-components/ComposeEmailWindow.svelte";
    import EmailContentsWindow from "$lib/my-components/EmailContentsWindow.svelte";
    import DesktopCommandBar from "$lib/my-components/DesktopCommandBar.svelte";
    import DesktopIcons from "$lib/my-components/DesktopIcons.svelte";

    import Progress from "$lib/components/ui/progress/progress.svelte";
    import LoginWindow from "$lib/my-components/LoginWindow.svelte";
    import ContactListWindow from "$lib/my-components/ContactListWindow.svelte";
    import CategoriesWindow from "$lib/my-components/CategoriesWindow.svelte";
    import TagsWindow from "$lib/my-components/TagsWindow.svelte";

    const registery: Record<string, ConstructorOfATypedSvelteComponent> = {
        [WindowType.EmailList.toString()]: EmailListWindow,
        [WindowType.ComposeEmail.toString()]: ComposeEmailWindow,
        [WindowType.EmailContents.toString()]: EmailContentsWindow,
        [WindowType.ContactList.toString()]: ContactListWindow,
        [WindowType.CategoryList.toString()]: CategoriesWindow,
        [WindowType.TagList.toString()]: TagsWindow,
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
