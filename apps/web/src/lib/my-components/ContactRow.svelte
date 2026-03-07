<script lang="ts">
    import type { IGooglePerson, IPersonInfo } from "$lib/models";
    import ShortenedEmailList from "./ShortenedEmailList.svelte";

    let {
        person,
    }: {
        person: IGooglePerson;
    } = $props();

    let gperson = $derived(person.person);

    let name = $derived(gperson.names?.[0].displayName ?? "Unknown");
    let photo = $derived.by(() => {
        if (!gperson.photos) {
            return null;
        }
        let lastUrl: string | null = null;
        for (const p of gperson.photos) {
            if (p.metadata?.primary ?? false) {
                return p.url;
            }
            if (!lastUrl && p.url) {
                lastUrl = p.url;
            }
        }
        return lastUrl;
    });
    let emails: IPersonInfo[] = $derived.by(() => {
        return (
            gperson.emailAddresses?.map((a) => {
                return {
                    email: a.value,
                    name: a.displayName ?? "",
                } as IPersonInfo;
            }) ?? []
        );
    });
    let phoneNumbers = $derived.by(() => {
        return (
            gperson.phoneNumbers?.map((a) => {
                return a.canonicalForm ?? a.value ?? "";
            }) ?? []
        );
    });
</script>

<div class="flex mb-1">
    <div class="rounded-full w-12 h-12 overflow-hidden mr-2">
        {#if photo}
            <img
                src={photo}
                alt={name}
                loading="lazy"
                referrerpolicy="no-referrer"
                decoding="async"
            />
        {/if}
    </div>
    <div class="grow-1">
        <div>{name}</div>
        <div class="text-sm">
            {#if emails.length > 0}
                <ShortenedEmailList contacts={emails} />
            {/if}
        </div>
        <div class="text-sm flex flex-wrap">
            {#if phoneNumbers.length > 0}
                {#each phoneNumbers as phoneNumber, idx}
                    <div>
                        {phoneNumber}
                        {#if idx < phoneNumbers.length - 1},
                        {/if}
                    </div>
                {/each}
            {/if}
        </div>
    </div>
</div>
