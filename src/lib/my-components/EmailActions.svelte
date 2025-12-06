<script lang="ts">
    import * as ButtonGroup from "$lib/components/ui/button-group/index.js";
    import { Button } from "$lib/components/ui/button/index.js";
    import { getContext, type Snippet } from "svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import ReplyIcon from "@lucide/svelte/icons/reply";
    import ReplyAllIcon from "@lucide/svelte/icons/reply-all";
    import ForwardIcon from "@lucide/svelte/icons/forward";
    import ArchiveIcon from "@lucide/svelte/icons/archive";
    import DeleteIcon from "@lucide/svelte/icons/delete";
    import MailIcon from "@lucide/svelte/icons/mail";
    import MailOpenIcon from "@lucide/svelte/icons/mail-open";
    import * as Tooltip from "$lib/components/ui/tooltip/index.js";
    import { emailMessageProvider } from "$lib/pods/EmailMessagePod";
    import {
        type IGmailEntry,
        WindowType,
        type IWindow,
        ComposeType,
    } from "$lib/models";

    let {
        email,
        pre,
    }: {
        email: IGmailEntry;
        pre?: Snippet;
    } = $props();

    const myWindow: IWindow = getContext("window");

    let isUnread = $derived(email?.labels?.includes("UNREAD") ?? false);

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
    function archive() {
        emailMessageProvider(email.messageId).archive();
    }
    function markAsUnread() {
        emailMessageProvider(email.messageId).markAsUnread();
    }
    function markAsRead() {
        emailMessageProvider(email.messageId).markAsRead();
    }
</script>

<div class="w-full p-2 flex">
    <ButtonGroup.Root>
        <ButtonGroup.Root>
            <Button variant="outline" onclick={archive}>
                <Tooltip.Provider>
                    <Tooltip.Root>
                        <Tooltip.Trigger>
                            <ArchiveIcon />
                        </Tooltip.Trigger>
                        <Tooltip.Content>Archive this email</Tooltip.Content>
                    </Tooltip.Root>
                </Tooltip.Provider>
            </Button>
            <Button variant="outline">
                <Tooltip.Provider>
                    <Tooltip.Root>
                        <Tooltip.Trigger>
                            <DeleteIcon />
                        </Tooltip.Trigger>
                        <Tooltip.Content>Delete this email</Tooltip.Content>
                    </Tooltip.Root>
                </Tooltip.Provider>
            </Button>
            {#if isUnread}
                <Button variant="outline" onclick={markAsRead}>
                    <Tooltip.Provider>
                        <Tooltip.Root>
                            <Tooltip.Trigger>
                                <MailOpenIcon />
                            </Tooltip.Trigger>
                            <Tooltip.Content>Mark as read</Tooltip.Content>
                        </Tooltip.Root>
                    </Tooltip.Provider>
                </Button>
            {:else}
                <Button variant="outline" onclick={markAsUnread}>
                    <Tooltip.Provider>
                        <Tooltip.Root>
                            <Tooltip.Trigger>
                                <MailIcon />
                            </Tooltip.Trigger>
                            <Tooltip.Content>Mark as unread</Tooltip.Content>
                        </Tooltip.Root>
                    </Tooltip.Provider>
                </Button>
            {/if}
        </ButtonGroup.Root>
        {@render pre?.()}
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
