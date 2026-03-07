<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import ComposeEmailContents from "$lib/my-components/ComposeEmailContents.svelte";
    import MailPlusIcon from "@lucide/svelte/icons/mail-plus";
    import { ComposeType } from "$lib/models";

    import type { IWindow, IGmailEntry, IComposeEmailMeta } from "$lib/models";
    import { emailContentsProvider } from "$lib/pods/EmailContentsPod";
    import Progress from "$lib/components/ui/progress/progress.svelte";
    import { emailMessageProvider } from "$lib/pods/EmailMessagePod";

    const {
        window,
        type,
        last,
        threadId,
    }: {
        window: IWindow;
        type: ComposeType;
        last: string | undefined;
        threadId: string | undefined;
    } = $props();

    const previousMessage = $derived(emailMessageProvider(last ?? ""));

    const previousContents = $derived(emailContentsProvider(last ?? ""));
    const loadingContents = $derived(previousContents.isLoading);

    const data: IComposeEmailMeta | null = $derived(
        getInitialData(type, $previousMessage),
    );

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

<Window {window} title={data?.subject ?? "Compose"} scrollable={false}>
    {#snippet windowTopLeft()}
        <MailPlusIcon />
    {/snippet}
    {#snippet content()}
        <div class="h-full">
            {#if !last && data}
                <ComposeEmailContents srcEmail={data} />
            {:else if $previousContents && data && !$loadingContents}
                <ComposeEmailContents
                    srcEmail={data}
                    previousContents={$previousContents}
                />
            {:else}
                <Progress value={null} />
            {/if}
        </div>
    {/snippet}
</Window>
