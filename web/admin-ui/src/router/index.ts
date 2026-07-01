import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/projects',
    },
    {
      path: '/projects',
      name: 'projects',
      component: () => import('@/views/ProjectList.vue'),
    },
    {
      path: '/projects/:id/:tab?',
      name: 'project-detail',
      component: () => import('@/views/ProjectDetail.vue'),
      props: true,
    },
    {
      path: '/tokens',
      name: 'tokens',
      component: () => import('@/views/TokenList.vue'),
    },
    {
      path: '/tokens/:id',
      name: 'token-detail',
      component: () => import('@/views/TokenDetail.vue'),
      props: true,
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFound.vue'),
    },
  ],
})

export default router
