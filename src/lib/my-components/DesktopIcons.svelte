<script lang="ts">
    import InboxIcon from "@lucide/svelte/icons/inbox";
    import StoreIcon from "@lucide/svelte/icons/store";
    import RefreshCcwIcon from "@lucide/svelte/icons/refresh-ccw";
    import ContactRoundIcon from "@lucide/svelte/icons/contact-round";
    import DesktopIcon from "$lib/my-components/DesktopIcon.svelte";
    import { windowProvider } from "$lib/pods/WindowsPod";
    import { WindowType, ComposeType } from "$lib/models";
    import { authHeaderProvider } from "$lib/pods/AuthPod";

    function openInbox() {
        windowProvider().open({
            type: WindowType.EmailList,
            props: {
                title: "Inbox",
            },
        });
    }
    function openPromotions() {
        windowProvider().open({
            type: WindowType.EmailList,
            props: {
                title: "Promotions",
                labels: ["CATEGORY_PROMOTIONS"],
            },
        });
    }
    async function doSync() {
        const headers = await authHeaderProvider().promise;
        await fetch("/api/people/sync", {
            headers,
        });
    }

    function openContacts() {
        windowProvider().open({
            type: WindowType.ContactList,
            props: {},
        });
    }
</script>

<DesktopIcon name="Inbox" onclick={openInbox}>
    {#snippet icon()}
        <InboxIcon class="size-12" />
    {/snippet}
</DesktopIcon>

<DesktopIcon name="Promotions" onclick={openPromotions}>
    {#snippet icon()}
        <StoreIcon class="size-12" />
    {/snippet}
</DesktopIcon>

<DesktopIcon name="Contacts" onclick={openContacts}>
    {#snippet icon()}
        <ContactRoundIcon class="size-12" />
    {/snippet}
</DesktopIcon>

<DesktopIcon name="Sync" onclick={doSync}>
    {#snippet icon()}
        <RefreshCcwIcon class="size-12" />
    {/snippet}
</DesktopIcon>
