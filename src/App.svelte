<script lang="ts">
    import "./app.css";
    import Counter from "$lib/Counter.svelte";
    import Window from "$lib/my-components/Window.svelte";
    import EmailRow from "$lib/my-components/EmailRow.svelte";
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
        <div slot="content"></div>
    </Window>

    <p></p>
</main>

<style>
</style>
