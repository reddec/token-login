import Home from './routes/Home.svelte';
import NotFound from './routes/NotFound.svelte';
import Token from './routes/Token.svelte';

export default {
    '/': Home,
    '/tokens/:id': Token,
    // The catch-all route must always be last
    '*': NotFound
};
