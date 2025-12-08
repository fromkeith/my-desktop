<script lang="ts">
    import InboxIcon from "@lucide/svelte/icons/inbox";
    import StoreIcon from "@lucide/svelte/icons/store";
    import RefreshCcwIcon from "@lucide/svelte/icons/refresh-ccw";
    import LibraryIcon from "@lucide/svelte/icons/library";
    import NewsPaperIcon from "@lucide/svelte/icons/newspaper";
    import TagsIcon from "@lucide/svelte/icons/tags";
    import ContactRoundIcon from "@lucide/svelte/icons/contact-round";
    import MessageCircleIcon from "@lucide/svelte/icons/message-circle";
    import DesktopIcon from "$lib/my-components/DesktopIcon.svelte";

    import { windowProvider } from "$lib/pods/WindowsPod";
    import { WindowType, ComposeType } from "$lib/models";
    import { authHeaderProvider } from "$lib/pods/AuthPod";

    function openInbox() {
        openWithLabels(
            [
                "-CATEGORY_PROMOTIONS",
                "-CATEGORY_UPDATES",
                "-CATEGORY_SOCIAL",
                "-SPAM",
                "INBOX",
            ],
            "Inbox",
        );
    }
    function openWithLabels(labels: string[], title: string) {
        windowProvider().open({
            type: WindowType.EmailList,
            props: {
                title,
                filter: {
                    labels,
                },
            },
        });
    }
    function openCategories() {
        windowProvider().open({
            type: WindowType.CategoryList,
            props: {},
        });
    }
    function openTags() {
        windowProvider().open({
            type: WindowType.TagList,
            props: {},
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

<DesktopIcon
    name="Promotions"
    onclick={() =>
        openWithLabels(["CATEGORY_PROMOTIONS", "-SPAM"], "Promotions")}
>
    {#snippet icon()}
        <StoreIcon class="size-12" />
    {/snippet}
</DesktopIcon>

<DesktopIcon
    name="Updates"
    onclick={() => openWithLabels(["CATEGORY_UPDATES", "-SPAM"], "Updates")}
>
    {#snippet icon()}
        <NewsPaperIcon class="size-12" />
    {/snippet}
</DesktopIcon>

<DesktopIcon
    name="Social"
    onclick={() => openWithLabels(["CATEGORY_SOCIAL", "-SPAM"], "Social")}
>
    {#snippet icon()}
        <MessageCircleIcon class="size-12" />
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

<DesktopIcon name="Categories" onclick={openCategories}>
    {#snippet icon()}
        <LibraryIcon class="size-12" />
    {/snippet}
</DesktopIcon>

<DesktopIcon name="Tags" onclick={openTags}>
    {#snippet icon()}
        <TagsIcon class="size-12" />
    {/snippet}
</DesktopIcon>
