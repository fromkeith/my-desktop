<script lang="ts">
    import { Separator } from "$lib/components/ui/separator/index.js";

    import {
        type IGmailEntry,
        WindowType,
        ComposeType,
        type IWindow,
    } from "$lib/models";

    import EmailThreadRow from "./EmailThreadRow.svelte";
    import { getContext } from "svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import ShortenedEmailList from "./ShortenedEmailList.svelte";

    export let threadId: string;
    export let thread: IGmailEntry[];
    export let openMessageId: string | undefined;

    let expanded: Set<string> = new Set();
    if (openMessageId) {
        expanded.add(openMessageId);
    }

    const myWindow: IWindow = getContext("window");
    $: last = thread.length > 0 ? thread[thread.length - 1] : null;
    $: lastMessageId = last?.messageId ?? null;

    function toggle(messageId: string) {
        if (expanded.has(messageId)) {
            expanded.delete(messageId);
        } else {
            expanded.add(messageId);
        }
        expanded = expanded;
    }
    function forward() {
        windowProvider().open(
            {
                type: WindowType.ComposeEmail,
                props: {
                    threadId: threadId,
                    last: lastMessageId,
                    type: ComposeType.Forward,
                },
            },
            myWindow,
        );
    }
    function reply() {
        windowProvider().open(
            {
                type: WindowType.ComposeEmail,
                props: {
                    threadId: threadId,
                    last: lastMessageId,
                    type: ComposeType.Reply,
                },
            },
            myWindow,
        );
    }
    function replyAll() {
        windowProvider().open(
            {
                type: WindowType.ComposeEmail,
                props: {
                    threadId: threadId,
                    last: lastMessageId,
                    type: ComposeType.ReplyAll,
                },
            },
            myWindow,
        );
    }
</script>

{#if thread.length > 0 && last}
    <h1 class="text-xs">{last.subject}</h1>
    <div class="mb-2 mr-2">
        <div class="text-md">
            <div class="w-full flex">
                <span class="font-bold mr-1">From</span>
                <ShortenedEmailList
                    contacts={[last.sender]}
                    hideCounter={true}
                    doClose={false}
                />
            </div>
        </div>
        <div class="text-sm">
            <div class="w-full flex">
                <span class="font-bold mr-1">To</span>
                <div class="grow">
                    <ShortenedEmailList
                        contacts={last.receiver}
                        doClose={false}
                    />
                </div>
            </div>
        </div>
        <Separator class="mt-2" />
        {#each thread as e (e.messageId)}
            <EmailThreadRow
                email={e}
                originalSubject={last.subject}
                expanded={expanded.has(e.messageId)}
                ontoggle={toggle}
            />
        {/each}
    </div>
{/if}
