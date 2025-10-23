<script lang="ts">
    import EmailRowActions from "$lib/my-components/EmailRowActions.svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import { WindowType, type IGmailEntry, type IWindow } from "$lib/models";
    import { getContext } from "svelte";
    import { dateFormat } from "$lib/pods/EmailListPod";

    export let isRead = false;
    export let email: IGmailEntry;
    let sender: string = email.sender.name || email.sender.email;
    let receivedAt = dateFormat.format(new Date(email.internalDate));

    let myWindow: IWindow = getContext("window");

    function openEmail() {
        windowProvider().open(
            {
                type: WindowType.EmailContents,
                props: {
                    email,
                },
            },
            myWindow,
        );
    }
</script>

<div class="flex w-full mb-2 items-center" class:opacity-70={isRead}>
    <EmailRowActions />
    <a href={'#'} class="flex-1 min-w-0" on:click|preventDefault={openEmail}>
        <div class="truncate text-xs text-blue-900">{sender}</div>
        <div class="truncate text-sm">{email.subject}</div>
        <div class="text-xs truncate opacity-70">{email.snippet}</div>
    </a>
    <div class="w-16 overflow-hidden text-right pv-1">
        {receivedAt}
    </div>
</div>
