<script lang="ts">
    import type { IPersonInfo } from "$lib/models";
    import ShortenedEmailList from "./ShortenedEmailList.svelte";
    import { createField } from "felte";

    let {
        contacts,
        name,
        errors,
        placeholder,
    }: {
        contacts: IPersonInfo[] | string;
        name: string;
        errors: string | undefined;
        placeholder: string | undefined;
    } = $props();

    const { field, onInput, onBlur } = createField(name, {
        defaultValue:
            typeof contacts === "string" ? contacts : JSON.stringify(contacts),
    });

    let contactsArray: IPersonInfo[] = $derived(
        typeof contacts === "string" ? JSON.parse(contacts) : contacts,
    );
    let errorsObj = $derived.by(() => {
        try {
            return errors ? JSON.parse(errors) : {};
        } catch (e) {
            return {};
        }
    });
    let errorsMap = $derived.by(() => {
        if (!errors || !errorsObj.path) {
            return new Map();
        }
        if (errorsObj.path.length === 1) {
            return new Map();
        }
        const index = errorsObj.path[1];
        return new Map([
            [
                index,
                {
                    tooltip: errorsObj.message,
                    class: "bg-red-500 text-white rounded p-1",
                },
            ],
        ]);
    });
    let generalError = $derived.by(() => {
        if (!errors) {
            return "";
        }
        if (errorsMap.size > 0) {
            return "";
        }
        return errors;
    });

    let textInput: HTMLInputElement;
    let currentInput: string = $state("");

    function focusText(e: MouseEvent) {
        textInput.focus();
    }

    function addContact(value: string) {
        const updated = [{ email: value, name: "" }, ...(contactsArray ?? [])];
        onInput(JSON.stringify(updated));
    }

    function inputKeyDown(e: KeyboardEvent) {
        // save and propagate
        if (e.key === "Tab") {
            const v = currentInput.trim();
            if (v.length > 0) {
                addContact(v);
            }
            currentInput = "";
        }
    }

    function inputKeyUp(e: KeyboardEvent) {
        // save and propagate
        if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            e.stopPropagation();
            const v = currentInput.trim();
            if (v.length > 0) {
                addContact(v);
            }
            currentInput = "";
        }
    }
    function inputBlur() {
        onBlur();
        const v = currentInput.trim();
        if (v.length > 0) {
            addContact(v);
        }
        currentInput = "";
    }
    function removeContact(e: IPersonInfo) {
        const updated = contactsArray.filter(
            (contact) => contact.email !== e.email,
        );
        onInput(JSON.stringify(updated));
    }
    function clickContact(contact: IPersonInfo, index: number) {
        removeContact(contact);
        currentInput = contact.email;
        textInput.focus();
    }
    //.
</script>

<div
    use:field
    {name}
    class="border-input bg-background selection:bg-primary dark:bg-input/30 selection:text-primary-foreground ring-offset-background placeholder:text-muted-foreground shadow-xs flex min-h-9 w-full min-w-0 rounded-md border px-3 py-1 text-base outline-none transition-[color,box-shadow] disabled:cursor-not-allowed disabled:opacity-50 md:text-sm focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive flex-wrap cursor-text"
    onclick={focusText}
>
    {#if (contactsArray?.length ?? 0) > 0}
        <div class="w-full">
            <ShortenedEmailList
                contacts={contactsArray}
                doClose={true}
                onremove={removeContact}
                highlight={errorsMap}
                {clickContact}
            />
        </div>
    {/if}
    <div class="flex w-full">
        <input
            bind:this={textInput}
            bind:value={currentInput}
            type="text"
            class="outline-0 grow-1"
            onkeydown={inputKeyDown}
            onkeyup={inputKeyUp}
            id={name}
            onblur={inputBlur}
            {placeholder}
        />
    </div>
    {#if generalError}
        <div class="text-red-500 text-sm">{generalError}</div>
    {/if}
</div>
