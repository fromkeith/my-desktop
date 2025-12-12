<script lang="ts">
    import "./app.css";

    import { windowListProvider } from "$lib/pods/WindowsPod";
    import { isAuthValidProvider } from "$lib/pods/AuthPod";
    import { WindowType, type IWindow } from "$lib/models";
    import { type Component } from "svelte";
    import type { Readable } from "svelte/store";

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

    const registry = {
        [WindowType.EmailList]: EmailListWindow,
        [WindowType.ComposeEmail]: ComposeEmailWindow,
        [WindowType.EmailContents]: EmailContentsWindow,
        [WindowType.ContactList]: ContactListWindow,
        [WindowType.CategoryList]: CategoriesWindow,
        [WindowType.TagList]: TagsWindow,
        [WindowType.LoginEmail]: LoginWindow,
    } satisfies Record<WindowType, Component<any>>;
    let windows: Readable<IWindow[]> = windowListProvider();
    let isAuthValid = isAuthValidProvider();
    let authLoading = $derived(isAuthValid.isLoading);
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
            {@const ComponentInst: Component<any> = registry[w.type]}
            <ComponentInst window={w} {...w.props} />
        {/each}
    {/if}
</main>

<style>
</style>
