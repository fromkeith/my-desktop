<script lang="ts">
    import * as Command from "$lib/components/ui/command/index.js";
    import CalendarIcon from "@lucide/svelte/icons/calendar";
    import MailsIcon from "@lucide/svelte/icons/mails";
    import MailPlusIcon from "@lucide/svelte/icons/mail-plus";
    import { WindowType, ComposeType } from "$lib/models";
    import { createDebounce } from "$lib/utils/debounce";
    import { windowProvider } from "$lib/pods/WindowsPod";
    let debounceClose = createDebounce();

    let toolbar: HTMLElement;

    let hasFocus: boolean = false;
    function focused() {
        hasFocus = true;
    }
    function blur() {
        debounceClose(200).then(() => {
            hasFocus = false;
        });
    }

    function getOffset() {
        return toolbar.getBoundingClientRect().bottom + 10;
    }

    function openEmails() {
        windowProvider().open({
            type: WindowType.EmailList,
            props: {},
            y: getOffset(),
        });
    }
    function openCompose() {
        windowProvider().open({
            type: WindowType.ComposeEmail,
            props: {
                type: ComposeType.New,
            },
            y: getOffset(),
        });
    }
</script>

<div
    class="absolute left-0 top-0 h-12 w-full bg-white flex justify-between"
    bind:this={toolbar}
    style="z-index: 9999999999;"
>
    <div></div>
    <div
        class:h-64={hasFocus}
        class:shadow-sm={hasFocus}
        class:border={hasFocus}
        class="rounded-lg w-64 transition-shadow"
    >
        <Command.Root>
            <Command.Input
                placeholder="Type a command"
                onfocus={focused}
                onblur={blur}
            />
            {#if hasFocus}
                <Command.List>
                    <Command.Empty>No results found.</Command.Empty>
                    <Command.Group heading="Suggestions">
                        <Command.Item onSelect={() => openEmails()}>
                            <MailsIcon />
                            <span>Emails</span>
                        </Command.Item>
                        <Command.Item onSelect={() => openCompose()}>
                            <MailPlusIcon />
                            <span>Compose</span>
                        </Command.Item>
                    </Command.Group>
                </Command.List>
            {/if}
        </Command.Root>
    </div>
    <div></div>
</div>
