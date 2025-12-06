<script lang="ts">
    import { type IGmailEntry } from "$lib/models";
    import { dateFormat } from "$lib/pods/EmailListPod";
    import EmailContentsDisplay from "./EmailContentsDisplay.svelte";
    import EmailContentsActions from "./EmailContentsActions.svelte";
    import { emailMessageProvider } from "$lib/pods/EmailMessagePod";

    let {
        email,
        originalSubject,
        expanded,
        ontoggle,
    }: {
        email: IGmailEntry;
        originalSubject: string;
        expanded: boolean;
        ontoggle: (messageId: string) => void;
    } = $props();

    let sender: string = $derived(email.sender.name || email.sender.email);
    let receivedAt = $derived(dateFormat.format(new Date(email.internalDate)));

    // mark as read when expanded
    $effect(() => {
        if (expanded && email.messageId) {
            emailMessageProvider(email.messageId).markAsRead();
        }
    });

    function toggleExpansion() {
        ontoggle(email.messageId);
    }
</script>

<div class="flex w-full mb-2 items-center">
    <a
        href={"#"}
        class="flex-1 min-w-0"
        on:click|preventDefault={toggleExpansion}
    >
        <div class="truncate text-xs text-blue-900">{sender}</div>
        {#if email.subject != originalSubject}
            <div class="truncate text-sm">{email.subject}</div>
        {/if}
        <div class="text-xs truncate opacity-70">{email.snippet}</div>
    </a>
    <div class="w-16 overflow-hidden text-right pv-1">
        {receivedAt}
    </div>
</div>
{#if expanded}
    <EmailContentsActions {email} />
    <EmailContentsDisplay messageId={email.messageId} />
{/if}
