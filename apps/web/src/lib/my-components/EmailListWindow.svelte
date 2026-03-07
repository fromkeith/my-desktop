<script lang="ts">
    import { threadListProvider } from "$lib/pods/ThreadListPod";
    import Window from "$lib/my-components/Window.svelte";
    import ThreadRow from "$lib/my-components/ThreadRow.svelte";
    import MailsIcon from "@lucide/svelte/icons/mails";

    import type { IWindow, IEmailListOptions } from "$lib/models";

    let {
        window,
        filter,
        title,
    }: {
        window: IWindow;
        filter?: IEmailListOptions;
        title?: string;
    } = $props();

    let threads = $derived(threadListProvider(filter));
</script>

<Window {window} {title}>
    {#snippet windowTopLeft()}
        <MailsIcon />
    {/snippet}
    {#snippet content()}
        {#each $threads as thread (thread.threadId)}
            <ThreadRow {thread} />
        {/each}
    {/snippet}
</Window>
