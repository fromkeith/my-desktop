<script lang="ts">
    import * as Tooltip from "$lib/components/ui/tooltip/index.js";
    import TagsIcon from "@lucide/svelte/icons/tags";
    import LibraryIcon from "@lucide/svelte/icons/library";
    import ListChecksIcon from "@lucide/svelte/icons/list-checks";
    import CircleIcon from "@lucide/svelte/icons/circle";
    import EmailActions from "$lib/my-components/EmailActions.svelte";
    import type { IGmailEntry } from "$lib/models";

    let {
        email,
    }: {
        email: IGmailEntry;
    } = $props();
</script>

<EmailActions {email}>
    {#snippet pre()}
        {#if email.tags.length > 0}
            <div class="p-1 pt-2">
                <Tooltip.Provider>
                    <Tooltip.Root>
                        <Tooltip.Trigger>
                            <TagsIcon size="16" />
                        </Tooltip.Trigger>
                        <Tooltip.Content>
                            Tags: {email.tags.join(", ")}
                        </Tooltip.Content>
                    </Tooltip.Root>
                </Tooltip.Provider>
            </div>
        {/if}
        {#if email.categories.length > 0}
            <div class="p-1 pt-2">
                <Tooltip.Provider>
                    <Tooltip.Root>
                        <Tooltip.Trigger>
                            <LibraryIcon size="16" />
                        </Tooltip.Trigger>
                        <Tooltip.Content>
                            Categories: {email.categories.join(", ")}
                        </Tooltip.Content>
                    </Tooltip.Root>
                </Tooltip.Provider>
            </div>
        {/if}
        {#if email.todos.length > 0}
            <div class="p-1 pt-2">
                <Tooltip.Provider>
                    <Tooltip.Root>
                        <Tooltip.Trigger>
                            <ListChecksIcon size="16" />
                        </Tooltip.Trigger>
                        <Tooltip.Content>
                            Todos:
                            {#each email.todos as todo}
                                <div class="flex">
                                    <CircleIcon size="16" class="pr-2" />
                                    {todo}
                                </div>
                            {/each}
                        </Tooltip.Content>
                    </Tooltip.Root>
                </Tooltip.Provider>
            </div>
        {/if}
    {/snippet}
</EmailActions>
