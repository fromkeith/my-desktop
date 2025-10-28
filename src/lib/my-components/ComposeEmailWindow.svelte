<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import ComposeEmailContents from "$lib/my-components/ComposeEmailContents.svelte";
    import MailPlusIcon from "@lucide/svelte/icons/mail-plus";
    import { ComposeType } from "$lib/models";

    import type { IWindow, IGmailEntry, IComposeEmailMeta } from "$lib/models";
    import { emailContentsProvider } from "$lib/pods/EmailContentsPod";
    import Progress from "$lib/components/ui/progress/progress.svelte";
    import { emailMessageProvider } from "$lib/pods/EmailMessagePod";

    export let window: IWindow;

    export let type: ComposeType;
    export let last: string | undefined;
    export let threadId: string | undefined;

    $: previousMessage = emailMessageProvider(last ?? "");

    $: previousContents = emailContentsProvider(last ?? "");
    $: loadingContents = previousContents.isLoading;

    $: data = getInitialData(type, $previousMessage);

    function getInitialData(
        type: ComposeType,
        previous: IGmailEntry | undefined,
    ): IComposeEmailMeta | null {
        if (type === ComposeType.New) {
            return {
                to: [],
                subject: "",
            };
        }
        if (!previous) {
            return null;
        }
        if (type === ComposeType.Reply) {
            return {
                to: [previous.replyTo ?? previous.sender],
                subject: `Re: ${previous?.subject ?? ""}`,
            };
        }
        if (type === ComposeType.ReplyAll) {
            return {
                to: [previous.replyTo ?? previous.sender],
                cc: (previous.additionalReceivers.cc ?? []).concat(
                    previous.receiver, // TODO: remove me
                ),
                subject: `Re: ${previous?.subject ?? ""}`,
            };
        }
        if (type === ComposeType.Forward) {
            return {
                to: [],
                subject: `Fwd: ${previous?.subject ?? ""}`,
            };
        }
        return {
            to: [],
        };
    }
</script>

<Window {window} scrollable={false}>
    <MailPlusIcon slot="window-top-left" />
    <div slot="content" class="h-full">
        {#if $previousContents && data && !$loadingContents}
            <ComposeEmailContents {data} previousContents={$previousContents} />
        {:else}
            <Progress />
        {/if}
    </div>
</Window>
