<script lang="ts">
    import { Separator } from "$lib/components/ui/separator/index.js";
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import ArchiveIcon from "@lucide/svelte/icons/archive";
    import DeleteIcon from "@lucide/svelte/icons/delete";
    import MailIcon from "@lucide/svelte/icons/mail";
    import ReplyIcon from "@lucide/svelte/icons/reply";
    import ReplyAllIcon from "@lucide/svelte/icons/reply-all";
    import ForwardIcon from "@lucide/svelte/icons/forward";
    import {
        type IGmailEntry,
        WindowType,
        ComposeType,
        type IWindow,
    } from "$lib/models";
    import { emailThreadProvider } from "$lib/pods/EmailThreadPod";
    import EmailThreadRow from "./EmailThreadRow.svelte";
    import { getContext } from "svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";

    export let email: IGmailEntry;

    $: thread = emailThreadProvider(email.threadId);

    let expanded: Set<string> = new Set();
    expanded.add(email.messageId);

    const myWindow: IWindow = getContext("window");

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
                    threadId: email.threadId,
                    last: email.messageId,
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
                    threadId: email.threadId,
                    last: email.messageId,
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
                    threadId: email.threadId,
                    last: email.messageId,
                    type: ComposeType.ReplyAll,
                },
            },
            myWindow,
        );
    }
</script>

<h1 class="text-lg">{email.subject}</h1>
<div class="mb-2">
    <div class="text-md">
        <span class="font-bold">From</span>
        <span>{email.sender.name} &lt;{email.sender.email}&gt;</span>
    </div>
    <div class="text-sm">
        <span class="font-bold">To</span>
        {#each email.receiver as rec, idx}
            <span class="ml-1"
                >{rec.name} &lt;{rec.email}&gt;
                {#if idx !== email.receiver.length - 1}
                    ,
                {/if}
            </span>
        {/each}
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
</div>
<Separator />
{#each $thread as e (e.messageId)}
    <EmailThreadRow
        email={e}
        originalSubject={email.subject}
        expanded={expanded.has(e.messageId)}
        on:toggle={toggle}
    />
{/each}
