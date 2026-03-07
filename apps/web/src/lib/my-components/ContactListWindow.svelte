<script lang="ts">
    import Window from "$lib/my-components/Window.svelte";
    import { peopleListProvider } from "$lib/pods/PeopleListPod";
    import type { IWindow } from "$lib/models";
    import ContactRow from "./ContactRow.svelte";

    const {
        window,
    }: {
        window: IWindow;
    } = $props();

    let people = peopleListProvider();
</script>

<Window {window} title="Contacts">
    {#snippet content()}
        <div>
            {#each $people as person (person.personId)}
                <ContactRow {person} />
            {/each}
        </div>
    {/snippet}
</Window>
