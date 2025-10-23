<script lang="ts">
    import EmailRowActions from "$lib/my-components/EmailRowActions.svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import { WindowType, type IGmailEntry, type IWindow } from "$lib/models";
    import { getContext } from "svelte";
    import {dateFormat} from "$lib/pods/EmailListPod";

    export let isRead = false;
    export let email: IGmailEntry;
    let sender: string = email.sender.name || email.sender.email;
    let receivedAt = dateFormat.format(new Date(email.internalDate));

    let myWindow: IWindow = getContext("window");

    function openEmail() {
        windowProvider().open(
            {
                type: WindowType.EmailContents,
                props: {},
            },
            myWindow,
        );
    }
</script>

<div class="flex w-full" class:opacity-70={isRead}>
    <EmailRowActions />
    <div class="flex flex-1 pv-1 min-w-0 items-center" on:click={openEmail}>
        <div class="w-32 overflow-hidden truncate" class:font-bold={!isRead}>
            {sender}
        </div>
        <div class="grow overflow-hidden min-w-0">
            <div class="truncate">{email.subject}</div>
            <div class="text-sm truncate">{email.snippet}</div>
        </div>
    </div>
    <div class="w-16 overflow-hidden text-right pv-1">
        {receivedAt}
    </div>
</div>
