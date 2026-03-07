<script lang="ts">
    import { type Snippet } from "svelte";
    import * as Menubar from "$lib/components/ui/menubar/index.js";
    import ListIcon from "@lucide/svelte/icons/list";
    import ListOrderedIcon from "@lucide/svelte/icons/list-ordered";
    import UndoIcon from "@lucide/svelte/icons/undo";
    import RedoIcon from "@lucide/svelte/icons/redo";
    import CheckIcon from "@lucide/svelte/icons/check";
    import { Button } from "$lib/components/ui/button/index.js";

    let {
        tipex,
        extras,
    }: {
        tipex: Tipex;
        extras?: Snippet;
    } = $props();

    let copySuccess = $state(false);

    async function copy() {
        try {
            await navigator.clipboard.writeText(tipex?.getText() || "");
            copySuccess = true;
            setTimeout(() => {
                copySuccess = false;
            }, 2000);
            tipex?.chain().focus().run();
        } catch (error) {
            console.error("Failed to copy:", error);
        }
    }
</script>

{#if tipex}
    <Menubar.Root>
        <Menubar.Menu>
            <Menubar.Trigger>A<span class="text-xs">A</span></Menubar.Trigger>
            <Menubar.Content>
                <Menubar.Item
                    class="text-lg font-bold"
                    onclick={() =>
                        tipex?.chain().focus().setHeading({ level: 1 }).run()}
                >
                    H1
                    {#if tipex?.isActive("heading", { level: 1 })}
                        <span class="grow"></span>
                        <span class="ml-2 text-xs">
                            <CheckIcon size="16" />
                        </span>
                    {/if}
                </Menubar.Item>
                <Menubar.Item
                    class="text-md font-bold"
                    onclick={() =>
                        tipex?.chain().focus().setHeading({ level: 2 }).run()}
                >
                    H2
                    {#if tipex?.isActive("heading", { level: 2 })}
                        <span class="grow"></span>
                        <span class="ml-2 text-xs">
                            <CheckIcon size="16" />
                        </span>
                    {/if}
                </Menubar.Item>
                <Menubar.Item
                    class="text-sm font-bold"
                    onclick={() =>
                        tipex?.chain().focus().setHeading({ level: 3 }).run()}
                >
                    H3
                    {#if tipex?.isActive("heading", { level: 3 })}
                        <span class="grow"></span>
                        <span class="ml-2 text-xs">
                            <CheckIcon size="16" />
                        </span>
                    {/if}
                </Menubar.Item>
                <Menubar.Item
                    onclick={() => tipex?.chain().focus().setParagraph().run()}
                >
                    Paragraph
                    {#if tipex?.isActive("paragraph")}
                        <span class="grow"></span>
                        <span class="ml-2 text-xs">
                            <CheckIcon size="16" />
                        </span>
                    {/if}
                </Menubar.Item>
            </Menubar.Content>
        </Menubar.Menu>
        <Menubar.Menu>
            <Menubar.Trigger>Style</Menubar.Trigger>
            <Menubar.Content>
                <Menubar.Item
                    class="text-bold"
                    onclick={() => tipex?.chain().focus().toggleBold().run()}
                >
                    Bold
                    {#if tipex?.isActive("bold")}
                        <span class="grow"></span>
                        <span class="ml-2 text-xs">
                            <CheckIcon size="16" />
                        </span>
                    {/if}
                </Menubar.Item>
                <Menubar.Item
                    class="italic"
                    onclick={() => tipex?.chain().focus().toggleItalic().run()}
                >
                    Italic
                    {#if tipex?.isActive("italic")}
                        <span class="grow"></span>
                        <span class="ml-2 text-xs">
                            <CheckIcon size="16" />
                        </span>
                    {/if}
                </Menubar.Item>
                <Menubar.Item
                    class="underline"
                    onclick={() =>
                        tipex?.chain().focus().toggleUnderline().run()}
                >
                    Underline
                    {#if tipex?.isActive("underline")}
                        <span class="grow"></span>
                        <span class="ml-2 text-xs">
                            <CheckIcon size="16" />
                        </span>
                    {/if}
                </Menubar.Item>
            </Menubar.Content>
        </Menubar.Menu>
        <Menubar.Menu>
            <Menubar.Trigger>List</Menubar.Trigger>
            <Menubar.Content>
                <Menubar.Item
                    onclick={() =>
                        tipex?.chain().focus().toggleBulletList().run()}
                >
                    <ListIcon />
                    {#if tipex?.isActive("bulletList")}
                        <span class="grow"></span>
                        <span class="ml-2 text-xs">
                            <CheckIcon size="16" />
                        </span>
                    {/if}
                </Menubar.Item>
                <Menubar.Item
                    onclick={() =>
                        tipex?.chain().focus().toggleOrderedList().run()}
                >
                    <ListOrderedIcon />
                    {#if tipex?.isActive("orderedList")}
                        <span class="grow"></span>
                        <span class="ml-2 text-xs">
                            <CheckIcon size="16" />
                        </span>
                    {/if}
                </Menubar.Item>
            </Menubar.Content>
        </Menubar.Menu>
        <span class="grow" />
        <button
            class="focus:bg-accent focus:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground outline-hidden flex select-none items-center rounded-sm px-2 py-1 text-sm font-medium cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            onclick={() => tipex?.chain().focus().undo().run()}
            disabled={!tipex?.can().undo()}
        >
            <UndoIcon size="16" />
        </button>

        <button
            class="focus:bg-accent focus:text-accent-foreground data-[state=open]:bg-accent data-[state=open]:text-accent-foreground outline-hidden flex select-none items-center rounded-sm px-2 py-1 text-sm font-medium cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            onclick={() => tipex?.chain().focus().redo().run()}
            disabled={!tipex?.can().redo()}
        >
            <RedoIcon size="16" />
        </button>
        {@render extras?.()}
    </Menubar.Root>
{/if}
