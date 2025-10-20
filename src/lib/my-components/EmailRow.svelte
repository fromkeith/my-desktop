<script lang="ts">
    import EmailRowActions from "$lib/my-components/EmailRowActions.svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import { WindowType } from "$lib/models";
    import { getContext } from "svelte";

    export let isRead = false;
    let sender = "Amazon.ca";
    let subject = 'Shipped: "Repel Umbrella Windproof..." and 3 more items';
    let preheader =
        "package was shipped!  Out for delivery Delivered Arriving tomorrow";
    let receivedAt = "Oct 17";

    let myWindow = getContext("window");

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
        <div class="w-32 overflow-hidden" class:font-bold={!isRead}>
            {sender}
        </div>
        <div class="grow overflow-hidden min-w-0">
            <div class="truncate">{subject}</div>
            <div class="text-sm truncate">{preheader}</div>
        </div>
    </div>
    <div class="w-16 overflow-hidden text-right pv-1">
        {receivedAt}
    </div>
</div>
