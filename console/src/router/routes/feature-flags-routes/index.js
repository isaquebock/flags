/** @type {import('vue-router').RouteRecordRaw} */
export const featureFlagsRoutes = {
  path: '/feature-flags',
  name: 'feature-flags',
  children: [
    {
      path: '',
      name: 'feature-flags-account',
      component: () => import('@views/FeatureFlags/AccountView.vue'),
      meta: {
        title: 'Feature Flags',
        breadCrumbs: [{ label: 'Feature Flags', to: '/feature-flags' }]
      }
    },
    {
      path: 'list',
      name: 'list-feature-flags',
      component: () => import('@views/FeatureFlags/ListView.vue'),
      meta: {
        title: 'Feature Flags',
        breadCrumbs: [{ label: 'Feature Flags', to: '/feature-flags/list' }]
      }
    },
    {
      path: 'create',
      name: 'create-feature-flags',
      component: () => import('@views/FeatureFlags/CreateView.vue'),
      meta: {
        title: 'Create Flag',
        breadCrumbs: [
          { label: 'Feature Flags', to: '/feature-flags/list' },
          { label: 'Create', to: '/feature-flags/create' }
        ]
      }
    },
    {
      path: 'edit/:id',
      name: 'edit-feature-flags',
      component: () => import('@views/FeatureFlags/EditView.vue'),
      meta: {
        title: 'Edit Flag',
        breadCrumbs: [
          { label: 'Feature Flags', to: '/feature-flags/list' },
          { label: 'Edit Flag', dynamic: true, routeParam: 'id' }
        ]
      }
    }
  ]
}
