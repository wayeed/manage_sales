import { defineStore } from 'pinia'
import Layout from '@/layout/index.vue'

/**
 * 静态路由配置（不需要权限即可访问）
 */
export const constantRoutes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { title: '登录', hidden: true },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Layout,
    meta: { title: '个人中心', hidden: true },
    children: [
      {
        path: '',
        name: 'UserProfile',
        component: () => import('@/views/profile/index.vue'),
        meta: { title: '个人中心' },
      },
    ],
  },
]

// 404 路由，必须放在动态路由最后添加
export const notFoundRoute = {
  path: '/:pathMatch(.*)*',
  name: 'NotFound',
  component: () => import('@/views/login/index.vue'),
  meta: { title: '404', hidden: true },
}

/**
 * 动态路由配置（根据权限加载）
 */
export const asyncRoutes = [
  // ==================== 仪表盘 ====================
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '仪表盘', icon: 'Odometer' },
      },
    ],
  },

  // ==================== 系统管理 ====================
  {
    path: '/system',
    component: Layout,
    name: 'System',
    meta: { title: '系统管理', icon: 'Setting' },
    children: [
      {
        path: 'user',
        name: 'UserManage',
        component: () => import('@/views/system/user/index.vue'),
        meta: { title: '用户管理', icon: 'User' },
      },
      {
        path: 'role',
        name: 'RoleManage',
        component: () => import('@/views/system/role/index.vue'),
        meta: { title: '角色管理', icon: 'Lock' },
      },
      {
        path: 'permission',
        name: 'PermissionManage',
        component: () => import('@/views/system/permission/index.vue'),
        meta: { title: '权限管理', icon: 'Key' },
      },
      {
        path: 'config',
        name: 'SystemConfig',
        component: () => import('@/views/system/config/index.vue'),
        meta: { title: '系统配置', icon: 'Tools' },
      },
      {
        path: 'log',
        name: 'SystemLog',
        component: () => import('@/views/system/log/index.vue'),
        meta: { title: '操作日志', icon: 'Notebook' },
      },
      {
        path: 'store',
        name: 'StoreManage',
        component: () => import('@/views/system/store/index.vue'),
        meta: { title: '门店管理', icon: 'Shop' },
      },
      {
        path: 'app-version',
        name: 'AppVersionManage',
        component: () => import('@/views/system/app-version/index.vue'),
        meta: { title: 'APP版本管理', icon: 'Monitor' },
      },
      {
        path: 'maintenance',
        name: 'SystemMaintenance',
        component: () => import('@/views/system/maintenance/index.vue'),
        meta: { title: '平台维护', icon: 'Tools' },
      },
    ],
  },

  // ==================== 商品管理 ====================
  {
    path: '/product',
    component: Layout,
    name: 'Product',
    meta: { title: '商品管理', icon: 'Goods' },
    children: [
      {
        path: 'category',
        name: 'ProductCategory',
        component: () => import('@/views/product/category/index.vue'),
        meta: { title: '商品分类', icon: 'Menu' },
      },
      {
        path: 'product',
        name: 'ProductList',
        component: () => import('@/views/product/product/index.vue'),
        meta: { title: '商品列表', icon: 'Goods' },
      },
      {
        path: 'product/add',
        name: 'ProductAdd',
        component: () => import('@/views/product/product/form.vue'),
        meta: { title: '新增商品', icon: 'Plus', hidden: true },
      },
      {
        path: 'product/edit',
        name: 'ProductEdit',
        component: () => import('@/views/product/product/form.vue'),
        meta: { title: '编辑商品', icon: 'Edit', hidden: true },
      },
    ],
  },

  // ==================== 订单管理 ====================
  {
    path: '/order',
    component: Layout,
    name: 'Order',
    meta: { title: '订单管理', icon: 'Document' },
    children: [
      {
        path: 'index',
        name: 'OrderList',
        component: () => import('@/views/order/index.vue'),
        meta: { title: '订单列表', icon: 'List' },
      },
      {
        path: 'detail/:id',
        name: 'OrderDetail',
        component: () => import('@/views/order/detail.vue'),
        meta: { title: '订单详情', icon: 'View', hidden: true },
      },
      {
        path: 'outbound-approval',
        name: 'OutboundApproval',
        component: () => import('@/views/order/outbound-approval.vue'),
        meta: { title: '出库审批', icon: 'CircleCheck' },
      },
    ],
  },

  // ==================== 库存管理 ====================
  {
    path: '/inventory',
    component: Layout,
    name: 'Inventory',
    meta: { title: '库存管理', icon: 'Box' },
    children: [
      {
        path: 'stock',
        name: 'InventoryStock',
        component: () => import('@/views/inventory/stock/index.vue'),
        meta: { title: '库存查询', icon: 'Search' },
      },
      {
        path: 'stocktaking',
        name: 'InventoryStocktaking',
        component: () => import('@/views/inventory/stocktaking/index.vue'),
        meta: { title: '库存盘点', icon: 'DocumentChecked' },
      },
      {
        path: 'purchase',
        name: 'InventoryPurchase',
        component: () => import('@/views/inventory/purchase/index.vue'),
        meta: { title: '采购管理', icon: 'ShoppingCart' },
      },
      {
        path: 'purchase/detail/:id',
        name: 'PurchaseDetail',
        component: () => import('@/views/inventory/purchase/detail.vue'),
        meta: { title: '采购详情', icon: 'View', hidden: true },
      },
      {
        path: 'receipt',
        name: 'InventoryReceipt',
        component: () => import('@/views/inventory/receipt/index.vue'),
        meta: { title: '回货管理', icon: 'Box' },
      },
      {
        path: 'receipt/detail/:id',
        name: 'ReceiptDetail',
        component: () => import('@/views/inventory/receipt/detail.vue'),
        meta: { title: '回货详情', icon: 'View', hidden: true },
      },
      {
        path: 'transfer',
        name: 'InventoryTransfer',
        component: () => import('@/views/inventory/transfer/index.vue'),
        meta: { title: '库存调拨', icon: 'Sort' },
      },
      {
        path: 'gift',
        name: 'InventoryGift',
        component: () => import('@/views/inventory/gift/index.vue'),
        meta: { title: '赠品管理', icon: 'Present' },
      },
      {
        path: 'supplier',
        name: 'InventorySupplier',
        component: () => import('@/views/inventory/supplier/index.vue'),
        meta: { title: '供应商管理', icon: 'OfficeBuilding' },
      },
      {
        path: 'warehouse',
        name: 'InventoryWarehouse',
        component: () => import('@/views/inventory/warehouse/index.vue'),
        meta: { title: '仓库管理', icon: 'House' },
      },
      {
        path: 'alert',
        name: 'InventoryAlert',
        component: () => import('@/views/inventory/alert/index.vue'),
        meta: { title: '库存预警', icon: 'Warning' },
      },
      {
        path: 'transaction',
        name: 'InventoryTransaction',
        component: () => import('@/views/inventory/transaction/index.vue'),
        meta: { title: '流水记录', icon: 'Tickets' },
      },
      {
        path: 'trace',
        name: 'InventoryTrace',
        component: () => import('@/views/inventory/trace/index.vue'),
        meta: { title: '库存穿透', icon: 'Connection' },
      },
    ],
  },

  // ==================== 送货管理 ====================
  {
    path: '/delivery',
    component: Layout,
    name: 'Delivery',
    meta: { title: '送货管理', icon: 'Van' },
    children: [
      {
        path: 'index',
        name: 'DeliveryList',
        component: () => import('@/views/delivery/index.vue'),
        meta: { title: '送货记录', icon: 'List' },
      },
      {
        path: 'pending',
        name: 'PendingDelivery',
        component: () => import('@/views/delivery/pending.vue'),
        meta: { title: '待送货订单', icon: 'Bell' },
      },
    ],
  },

  // ==================== 提成管理 ====================
  {
    path: '/commission',
    component: Layout,
    name: 'Commission',
    meta: { title: '提成管理', icon: 'Money' },
    children: [
      {
        path: 'detail',
        name: 'CommissionDetail',
        component: () => import('@/views/commission/detail/index.vue'),
        meta: { title: '提成明细', icon: 'Document' },
      },
      {
        path: 'summary',
        name: 'CommissionSummary',
        component: () => import('@/views/commission/summary/index.vue'),
        meta: { title: '提成汇总', icon: 'DataLine' },
      },
      {
        path: 'fund',
        name: 'CommissionFund',
        component: () => import('@/views/commission/fund/index.vue'),
        meta: { title: '提成基金', icon: 'Wallet' },
      },
      {
        path: 'referral',
        name: 'CommissionReferral',
        component: () => import('@/views/commission/referral/index.vue'),
        meta: { title: '老带新管理', icon: 'Connection' },
      },
      {
        path: 'settlement',
        name: 'CommissionSettlement',
        component: () => import('@/views/commission/settlement/index.vue'),
        meta: { title: '结算管理', icon: 'Timer' },
      },
    ],
  },

  // ==================== 工资管理 ====================
  {
    path: '/salary',
    component: Layout,
    name: 'Salary',
    meta: { title: '工资管理', icon: 'CreditCard' },
    children: [
      {
        path: 'index',
        name: 'SalaryList',
        component: () => import('@/views/salary/index.vue'),
        meta: { title: '工资列表', icon: 'List' },
      },
      {
        path: 'detail/:id',
        name: 'SalaryDetail',
        component: () => import('@/views/salary/detail.vue'),
        meta: { title: '工资详情', icon: 'View', hidden: true },
      },
    ],
  },

  // ==================== 数据分析 ====================
  {
    path: '/report',
    component: Layout,
    name: 'Report',
    meta: { title: '数据分析', icon: 'DataAnalysis' },
    children: [
      {
        path: 'sales',
        name: 'SalesReport',
        component: () => import('@/views/report/sales/index.vue'),
        meta: { title: '销售报表', icon: 'TrendCharts' },
      },
      {
        path: 'ranking',
        name: 'RankingReport',
        component: () => import('@/views/report/ranking/index.vue'),
        meta: { title: '销售排行', icon: 'Trophy' },
      },
      {
        path: 'profit',
        name: 'ProfitReport',
        component: () => import('@/views/report/profit/index.vue'),
        meta: { title: '利润报表', icon: 'Coin' },
      },
      {
        path: 'payment',
        name: 'PaymentReport',
        component: () => import('@/views/report/payment/index.vue'),
        meta: { title: '收款报表', icon: 'CreditCard' },
      },
      {
        path: 'inventory',
        name: 'InventoryReport',
        component: () => import('@/views/report/inventory/index.vue'),
        meta: { title: '库存报表', icon: 'Box' },
      },
      {
        path: 'commission',
        name: 'CommissionReport',
        component: () => import('@/views/report/commission/index.vue'),
        meta: { title: '提成报表', icon: 'PieChart' },
      },
    ],
  },

  // ==================== 客户管理 ====================
  {
    path: '/customer',
    component: Layout,
    name: 'Customer',
    meta: { title: '客户管理', icon: 'Avatar' },
    children: [
      {
        path: 'index',
        name: 'CustomerList',
        component: () => import('@/views/customer/index.vue'),
        meta: { title: '客户列表', icon: 'UserFilled' },
      },
    ],
  },

  // ==================== 同行管理 ====================
  {
    path: '/peer',
    component: Layout,
    name: 'Peer',
    meta: { title: '同行管理', icon: 'User' },
    children: [
      {
        path: 'index',
        name: 'PeerList',
        component: () => import('@/views/peer/index.vue'),
        meta: { title: '同行列表', icon: 'UserFilled' },
      },
    ],
  },
]

export const usePermissionStore = defineStore('permission', {
  state: () => ({
    routes: [],
    menuList: [],
  }),

  getters: {
    /**
     * 获取用于显示的菜单列表（过滤掉隐藏的路由）
     */
    visibleMenuList: (state) => {
      return state.menuList.filter((route) => !route.meta?.hidden)
    },
  },

  actions: {
    /**
     * 根据用户角色生成可访问的路由
     * @param {Array} roles - 用户角色列表
     * @returns {Array} 可访问的路由列表
     */
    generateRoutes(roles) {
      let accessedRoutes
      if (roles.includes('admin')) {
        // 管理员拥有所有权限
        accessedRoutes = asyncRoutes
      } else {
        // 非管理员根据 roles 过滤路由
        accessedRoutes = filterAsyncRoutes(asyncRoutes, roles)
      }
      // 404 路由必须放在最后，确保动态路由优先匹配
      accessedRoutes = accessedRoutes.concat(notFoundRoute)
      this.routes = constantRoutes.concat(accessedRoutes)
      // 过滤隐藏路由生成菜单
      this.menuList = buildMenuList(accessedRoutes)
      return accessedRoutes
    },

    /**
     * 重置路由状态
     */
    resetRoutes() {
      this.routes = []
      this.menuList = []
    },
  },
})

/**
 * 递归过滤异步路由
 * @param {Array} routes - 异步路由表
 * @param {Array} roles - 用户角色列表
 */
function filterAsyncRoutes(routes, roles) {
  const result = []
  routes.forEach((route) => {
    const tmp = { ...route }
    if (hasPermission(roles, tmp)) {
      if (tmp.children) {
        tmp.children = filterAsyncRoutes(tmp.children, roles)
      }
      result.push(tmp)
    }
  })
  return result
}

/**
 * 判断是否有路由权限
 */
function hasPermission(roles, route) {
  if (route.meta?.roles) {
    return roles.some((role) => route.meta.roles.includes(role))
  }
  return true
}

/**
 * 构建菜单列表（过滤隐藏项，处理单子菜单）
 * @param {Array} routes
 * @returns {Array}
 */
function buildMenuList(routes) {
  return routes
    .filter((route) => !route.meta?.hidden)
    .map((route) => {
      const menu = { ...route }
      if (menu.children) {
        menu.children = menu.children.filter((child) => !child.meta?.hidden)
      }
      return menu
    })
}
