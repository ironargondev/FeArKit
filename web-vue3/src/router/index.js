import { createRouter, createWebHashHistory } from 'vue-router';

/**
 * Build the app router.
 * @param {(path: string) => Promise} loadSFC - factory that resolves a .vue path to a component
 */
export function createAppRouter(loadSFC) {
  return createRouter({
    history: createWebHashHistory(),
    routes: [
      { path: '/',                component: () => loadSFC('src/views/Home.vue') },
      { path: '/:pathMatch(.*)*', component: () => loadSFC('src/views/NotFound.vue') },
    ],
  });
}
