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
        component: () => import('../views/ArticleEdit.vue'),
        meta: { layout: 'editor' }
      },
      {
        path: 'articles/edit/:id',
        name: 'ArticleEdit',
        component: () => import('../views/ArticleEdit.vue'),
        meta: { layout: 'editor' }
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
      },
      {
        path: ':legacySettingsTab(site|content|seo|storage|ai|security)',
        redirect: (to) => ({
          name: 'Settings',
          query: { tab: to.params.legacySettingsTab }
        })
      }
    ]
  }
];

const router = createRouter({
  history: createWebHashHistory(),
  routes
});

/**
 * 规范化登录回跳地址，避免跳转到站外地址或登录页自身。
 */
const resolveRedirectPath = (redirect) => {
  if (typeof redirect !== 'string' || !redirect.startsWith('/')) {
    return '/';
  }

  return redirect === '/login' ? '/' : redirect;
};

router.beforeEach(async (to, _from, next) => {
  const userStore = useUserStore();
  
  if (to.meta.requiresAuth) {
    if (!userStore.authed) {
      await userStore.loadMe();
      if (!userStore.authed) {
        next({
          path: '/login',
          query: { redirect: to.fullPath }
        });
      } else {
        next();
      }
    } else {
      next();
    }
  } else if (to.path === '/login' && userStore.authed) {
    next(resolveRedirectPath(to.query.redirect));
  } else {
    next();
  }
});

export default router;
