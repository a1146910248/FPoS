import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '@/components/DashboardPanel.vue'
import Transaction from '@/components/TransactionPanel.vue'
import BlockDetail from '@/components/BlockDetail.vue'
import TransactionDetail from '@/components/TransactionDetail.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Dashboard,
    },
    {
      path: '/transaction',
      name: 'transaction',
      component: Transaction,
    },
    {
      path: '/blockDetail/:block_hash',
      name: 'blockDetail',
      component: BlockDetail,
      props: true,
    },
    {
      path: '/transactionDetail/:hash',
      name: 'transactionDetail',
      component: TransactionDetail,
      props: true,
    },
  ],
})

export default router
