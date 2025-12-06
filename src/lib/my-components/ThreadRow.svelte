<script lang="ts">
    import EmailRowActions from "$lib/my-components/EmailRowActions.svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import { WindowType, type IThread, type IWindow } from "$lib/models";
    import { getContext } from "svelte";
    import { dateFormat } from "$lib/pods/EmailListPod";
    import EmailLabels from "./EmailLabels.svelte";

    let {
        thread,
    }: {
        thread: IThread;
    } = $props();
    let receivedAt = dateFormat.format(
        new Date(thread.mostRecentInternalDate ?? 0),
    );

    let mostRecentMessage = $derived.by(() => {
        return (
            thread.messages.find(
                (m) => m.internalDate >= thread.mostRecentInternalDate,
            ) ?? thread.messages[0]
        );
    });

    let sender = $derived(
        mostRecentMessage.sender?.name || mostRecentMessage.sender!.email,
    );

    let myWindow: IWindow = getContext("window");

    function openEmail() {
        windowProvider().open(
            {
                type: WindowType.EmailContents,
                props: {
                    threadId: thread.threadId,
                    openTo: mostRecentMessage.messageId,
                },
            },
            myWindow,
        );
    }
    let isRead = $derived.by(() => {
        for (const m of thread.messages) {
            if (m.labels?.indexOf("UNREAD") !== -1) {
                return false;
            }
        }
        return true;
    });
    let labels = $derived.by(() => {
        const labelSet = new Set<string>();
        for (const m of thread.messages) {
            if (m.labels) {
                for (const label of m.labels) {
                    labelSet.add(label);
                }
            }
        }
        return Array.from(labelSet);
    });
</script>

<div class="flex w-full mb-2 items-center" class:opacity-70={isRead}>
    <EmailRowActions />
    <a href={"#"} class="flex-1 min-w-0" on:click|preventDefault={openEmail}>
        <div class="w-full overflow-hidden flex">
            <EmailLabels {labels} />
            <div class="truncate text-xs text-blue-900 grow-1">
                {#if thread.messages.length > 1}
                    ({thread.messages.length})
                {/if}
                {sender}
            </div>
        </div>
        <div class="truncate text-sm">{mostRecentMessage.subject}</div>
        <div class="text-xs truncate opacity-70">
            {mostRecentMessage.snippet}
        </div>
    </a>
    <div class="w-16 overflow-hidden text-right pv-1">
        {receivedAt}
    </div>
</div>
