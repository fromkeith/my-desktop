<script lang="ts">
    import { Separator } from "$lib/components/ui/separator/index.js";
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import ArchiveIcon from "@lucide/svelte/icons/archive";
    import DeleteIcon from "@lucide/svelte/icons/delete";
    import MailIcon from "@lucide/svelte/icons/mail";
    import ReplyIcon from "@lucide/svelte/icons/reply";
    import ChevronsUpDownIcon from "@lucide/svelte/icons/chevrons-up-down";
    import ReplyAllIcon from "@lucide/svelte/icons/reply-all";
    import ForwardIcon from "@lucide/svelte/icons/forward";
    import {
        type IGmailEntry,
        WindowType,
        ComposeType,
        type IWindow,
    } from "$lib/models";

    import EmailThreadRow from "./EmailThreadRow.svelte";
    import { getContext } from "svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import * as Collapsible from "$lib/components/ui/collapsible/index.js";
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

    function toggle(e: CustomEvent<string>) {
        if (expanded.has(e.detail)) {
            expanded.delete(e.detail);
        } else {
            expanded.add(e.detail);
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

{last}

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
        <div class="mt-1 flex flex-wrap">
            <ButtonGroup.Root>
                <Button variant="outline">
                    <ArchiveIcon />
                </Button>
                <Button variant="outline">
                    <DeleteIcon />
                </Button>
                <Button variant="outline">
                    <MailIcon />
                </Button>
            </ButtonGroup.Root>
            <div class="grow"></div>
            <Button variant="outline" onclick={forward}>
                <ForwardIcon />
            </Button>
            <div class="ml-1"></div>
            <ButtonGroup.Root>
                <Button variant="outline" onclick={reply}>
                    <ReplyIcon />
                </Button>
                <Button variant="outline" onclick={replyAll}>
                    <ReplyAllIcon />
                </Button>
            </ButtonGroup.Root>
        </div>
        <Separator />
        {#each thread as e (e.messageId)}
            <EmailThreadRow
                email={e}
                originalSubject={last.subject}
                expanded={expanded.has(e.messageId)}
                on:toggle={toggle}
            />
        {/each}
    </div>
{/if}
