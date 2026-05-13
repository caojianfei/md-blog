import { createRouter, createWebHashHistory } from 'vue-router';
import { useUserStore } from '../store/user';
import Layout from '../components/Layout.vue';

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue')
      },
      {
        path: 'articles',
        name: 'ArticleList',
        component: () => import('../views/ArticleList.vue')
      },
      {
        path: 'articles/new',
        name: 'ArticleNew',
        component: () => import('../views/ArticleEdit.vue')
      },
      {
        path: 'articles/edit/:id',
        name: 'ArticleEdit',
        component: () => import('../views/ArticleEdit.vue')
      },
      {
        path: 'categories',
        name: 'CategoryList',
        component: () => import('../views/CategoryList.vue')
      },
      {
        path: 'tags',
        name: 'TagList',
        component: () => import('../views/TagList.vue')
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('../views/Settings.vue')
      }
    ]
  }
];

const router = createRouter({
  history: createWebHashHistory(),
  routes
});

router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore();
  
  if (to.meta.requiresAuth) {
    if (!userStore.authed) {
      await userStore.loadMe();
      if (!userStore.authed) {
        next('/login');
      } else {
        next();
      }
    } else {
      next();
    }
  } else if (to.path === '/login' && userStore.authed) {
    next('/');
  } else {
    next();
  }
});

export default router;
