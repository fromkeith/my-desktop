<script lang="ts">
    import { type IPersonInfo } from "$lib/models";
    import ShortenedEmail from "$lib/my-components/ShortenedEmail.svelte";
    import UsersRoundIcon from "@lucide/svelte/icons/users-round";
    import ChevronUpIcon from "@lucide/svelte/icons/chevron-up";
    import EmailContact from "./EmailContact.svelte";
    const {
        contacts,
        doCopy = true,
        doClose = false,
        onremove,
        highlight = new Map(),
        hideCounter = false,
    }: {
        contacts: IPersonInfo[];
        doCopy?: boolean;
        doClose?: boolean;
        onremove?: (c: IPersonInfo) => void;
        highlight?: Map<number, { tooltip: string; class: string }>;
        hideCounter?: boolean;
    } = $props();

    let expanded = $state(false);

    function toggleExpanded() {
        expanded = !expanded;
        console.log("toggleExpanded");
    }
</script>

{#if !expanded && contacts.length > 0}
    <button
        class="flex w-full justify-between cursor-pointer"
        onclick={toggleExpanded}
    >
        <div class="overflow-hidden text-ellipsis whitespace-nowrap">
            {#each contacts as rec, idx}
                <ShortenedEmail
                    contact={rec}
                    highlight={highlight.get(idx)}
                />{#if idx < contacts.length - 1},&nbsp;
                {/if}
            {/each}
        </div>
        {#if !hideCounter}
            <div class="text-sm w-16 flex justify-end">
                {contacts.length}
                <UsersRoundIcon class="size-4 ml-1" />
            </div>
        {/if}
    </button>
{:else}
    <div class="flex w-full justify-between items-start">
        <div>
            {#each contacts as contact, idx}
                <EmailContact
                    {doClose}
                    {doCopy}
                    {contact}
                    {onremove}
                    highlight={highlight.get(idx)}
                />
            {/each}
        </div>
        <button onclick={toggleExpanded} class="cursor-pointer">
            <div class="text-sm w-16 flex justify-end">
                <ChevronUpIcon class="size-4 ml-1" />
            </div>
        </button>
    </div>
{/if}
