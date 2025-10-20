<script lang="ts">
    import "./app.css";
    import Counter from "$lib/Counter.svelte";
    import Window from "$lib/my-components/Window.svelte";
    import EmailRow from "$lib/my-components/EmailRow.svelte";
    import EmailContents from "$lib/my-components/EmailContents.svelte";
    import ComposeEmailContents from "$lib/my-components/ComposeEmailContents.svelte";
</script>

<main class="w-screen h-screen">
    <div class="card">
        <Counter />
    </div>

    <Window>
        <div slot="content">
            <EmailRow isRead={false} />
            <EmailRow isRead={true} />
            <EmailRow isRead={true} />
            <EmailRow />
            <EmailRow isRead={true} />
        </div>
    </Window>
    <Window>
        <EmailContents slot="content" />
    </Window>
    <Window scrollable={false}>
        <ComposeEmailContents slot="content" />
    </Window>

    <p></p>
</main>

<style>
</style>
