<script lang="ts">
    import EmailRowActions from "$lib/my-components/EmailRowActions.svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import { WindowType, type IGmailEntry, type IWindow } from "$lib/models";
    import { createEventDispatcher, getContext } from "svelte";
    import { dateFormat } from "$lib/pods/EmailListPod";
    import EmailContentsDisplay from "./EmailContentsDisplay.svelte";

    const dispatch = createEventDispatcher();

    export let isRead = false;
    export let email: IGmailEntry;
    export let originalSubject: string;
    export let expanded: boolean;
    let sender: string = email.sender.name || email.sender.email;
    let receivedAt = dateFormat.format(new Date(email.internalDate));

    let myWindow: IWindow = getContext("window");

    function toggleExpansion() {
        dispatch("toggle", email.messageId);
    }
</script>

<div class="flex w-full mb-2 items-center">
    <a href={'#'} class="flex-1 min-w-0" on:click|preventDefault={toggleExpansion}>
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
    <EmailContentsDisplay messageId={email.messageId} />
{/if}
