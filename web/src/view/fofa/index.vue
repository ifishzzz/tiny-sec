<template>
  <div>
    <WarningBar :title="warningText" />

    <div class="gva-search-box">
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="searchInfo"
        :rules="searchRules"
        class="demo-form-inline"
      >
        <el-form-item label="搜索引擎" prop="engine" class="!mr-4">
          <el-radio-group v-model="searchInfo.engine">
            <el-radio-button label="fofa">FOFA</el-radio-button>
            <el-radio-button label="quake">Quake</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="查询语法" prop="query" class="!mr-4">
          <el-input
            v-model="searchInfo.query"
            class="!w-[460px]"
            clearable
            :placeholder="queryPlaceholder"
            @keydown.enter.prevent="onSubmit"
          />
        </el-form-item>
        <el-form-item label="全量搜索" prop="full">
          <el-switch v-model="searchInfo.full" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" :loading="loading" @click="onSubmit">
            查询
          </el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box !pt-0">
      <div class="rounded-md border border-gray-200 bg-white p-4">
        <div class="mb-3 flex items-center justify-between">
          <div class="text-base font-medium text-gray-800">{{ ruleTips.title }}</div>
          <span class="text-sm text-gray-500">{{ ruleTips.summary }}</span>
        </div>
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
          <div>
            <div class="mb-2 text-sm font-medium text-gray-700">常用规则</div>
            <div class="flex flex-wrap gap-2">
              <el-tag
                v-for="item in ruleTips.operators"
                :key="item"
                type="info"
                effect="plain"
              >
                {{ item }}
              </el-tag>
            </div>
          </div>
          <div>
            <div class="mb-2 text-sm font-medium text-gray-700">常用字段</div>
            <div class="flex flex-wrap gap-2">
              <el-tag
                v-for="item in ruleTips.fields"
                :key="item"
                effect="plain"
              >
                {{ item }}
              </el-tag>
            </div>
          </div>
          <div>
            <div class="mb-2 text-sm font-medium text-gray-700">查询示例</div>
            <div class="space-y-2">
              <div
                v-for="item in ruleTips.examples"
                :key="item"
                class="rounded bg-gray-50 px-3 py-2 font-mono text-xs text-gray-700"
              >
                {{ item }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <span class="text-sm text-gray-500">
          {{ helperText }}
        </span>
      </div>

      <el-table
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        :row-key="getRowKey"
      >
        <el-table-column align="left" label="Host" min-width="240">
          <template #default="{ row }">
            <a
              v-if="buildLink(row)"
              :href="buildLink(row)"
              target="_blank"
              rel="noopener noreferrer"
              class="text-primary hover:underline"
            >
              {{ row.host || '-' }}
            </a>
            <span v-else>{{ row.host || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="IP" prop="ip" min-width="140" />
        <el-table-column align="left" label="端口" prop="port" width="90" />
        <el-table-column align="left" label="协议" prop="protocol" width="100" />
        <el-table-column align="left" label="标题" prop="title" min-width="220" show-overflow-tooltip />
        <el-table-column align="left" label="域名" prop="domain" min-width="180" show-overflow-tooltip />
        <el-table-column align="left" label="服务" prop="server" min-width="180" show-overflow-tooltip />
        <el-table-column align="left" label="地区" min-width="140">
          <template #default="{ row }">
            {{ formatRegion(row) }}
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          layout="total, sizes, prev, pager, next, jumper"
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="pageSizeOptions"
          :total="total"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import WarningBar from '@/components/warningBar/warningBar.vue'
import { spaceSearch } from '@/api/fofa'

defineOptions({
  name: 'SpaceSearch'
})

const searchFormRef = ref()
const loading = ref(false)
const page = ref(1)
const pageSize = ref(10)
const total = ref(0)
const tableData = ref([])

const searchInfo = reactive({
  engine: 'fofa',
  query: '',
  full: false
})

const searchRules = reactive({
  query: [
    {
      required: true,
      message: '请输入查询语法',
      trigger: ['blur', 'change']
    },
    {
      validator: (_rule, value, callback) => {
        if (!value || !value.trim()) {
          callback(new Error('请输入查询语法'))
          return
        }
        callback()
      },
      trigger: ['blur', 'change']
    }
  ]
})

const warningText = computed(() => {
  if (searchInfo.engine === 'quake') {
    return '请先在参数管理中配置 `quake_key`，查询请求会由后端代理发送到 Quake。'
  }
  return '请先在参数管理中配置 `fofa_email` 和 `fofa_key`，查询请求会由后端代理发送到 FOFA。'
})

const queryPlaceholder = computed(() => {
  if (searchInfo.engine === 'quake') {
    return '例如：service:\"http\" AND country:\"China\"'
  }
  return '例如：app=\"nginx\" && country=\"CN\"'
})

const helperText = computed(() => {
  if (searchInfo.engine === 'quake') {
    return '支持标准 Quake 查询语法，关闭全量搜索时优先查询最新资产。'
  }
  return '支持标准 FOFA 查询语法，关闭全量搜索时优先查询最新资产。'
})

const ruleTips = computed(() => {
  if (searchInfo.engine === 'quake') {
    return {
      title: 'Quake 搜索规则提示',
      summary: '推荐使用英文逻辑运算符和字段查询。',
      operators: [
        'AND 逻辑与',
        'OR 逻辑或',
        'NOT 逻辑非',
        'field:"value" 精确/短语匹配',
        'port:443 数值匹配'
      ],
      fields: [
        'service:"http"',
        'service.product:"nginx"',
        'ip:"1.1.1.1"',
        'port:443',
        'country:"China"',
        'city:"Beijing"',
        'title:"后台管理"'
      ],
      examples: [
        'service:"http" AND port:443',
        'service.product:"nginx" AND country:"China"',
        'title:"login" AND country:"China" AND port:8443'
      ]
    }
  }

  return {
    title: 'FOFA 搜索规则提示',
    summary: '推荐使用双引号包裹字符串，并使用 && / || 组合条件。',
    operators: [
      '&& 逻辑与',
      '|| 逻辑或',
      '== / = 字段匹配',
      '字段值建议用双引号',
      '支持括号组合条件'
    ],
    fields: [
      'app="nginx"',
      'title="后台管理"',
      'country="CN"',
      'city="Beijing"',
      'host="example.com"',
      'ip="1.1.1.1"',
      'port="443"'
    ],
    examples: [
      'app="nginx" && country="CN"',
      'title="login" && port="443"',
      '(app="nginx" || app="Apache") && country="CN"'
    ]
  }
})

const pageSizeOptions = computed(() => {
  return searchInfo.engine === 'quake' ? [10, 20, 50, 100] : [10, 20, 50, 100]
})

const getTableData = async () => {
  if (!searchInfo.query.trim()) {
    tableData.value = []
    total.value = 0
    return
  }

  loading.value = true
  try {
    const res = await spaceSearch({
      engine: searchInfo.engine,
      query: searchInfo.query.trim(),
      full: searchInfo.full,
      page: page.value,
      pageSize: pageSize.value
    })

    if (res.code === 0) {
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
      page.value = res.data.page || page.value
      pageSize.value = res.data.pageSize || pageSize.value
      return
    }

    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const onSubmit = () => {
  searchFormRef.value?.validate(async (valid) => {
    if (!valid) return
    page.value = 1
    await getTableData()
  })
}

const onReset = () => {
  searchInfo.engine = 'fofa'
  searchInfo.query = ''
  searchInfo.full = false
  page.value = 1
  pageSize.value = 10
  total.value = 0
  tableData.value = []
  searchFormRef.value?.clearValidate()
}

const handleCurrentChange = async (value) => {
  page.value = value
  await getTableData()
}

const handleSizeChange = async (value) => {
  pageSize.value = value
  page.value = 1
  await getTableData()
}

const getRowKey = (row) => {
  return row.host || `${row.ip || ''}:${row.port || ''}:${row.title || ''}`
}

const buildLink = (row) => {
  if (!row.host) {
    return ''
  }

  if (row.host.startsWith('http://') || row.host.startsWith('https://')) {
    return row.host
  }

  if (row.protocol) {
    return `${row.protocol}://${row.host}`
  }

  return ''
}

const formatRegion = (row) => {
  return [row.country, row.city].filter(Boolean).join(' / ') || '-'
}
</script>
