<template>
  <div class="mongodb-console">
    <aside class="console-sidebar" :class="{ 'console-sidebar--collapsed': sidebarCollapsed }">
      <div class="sidebar-header" :class="{ 'sidebar-header--collapsed': sidebarCollapsed }">
        <button class="sidebar-toggle sidebar-toggle--header" :class="{ 'sidebar-toggle--collapsed': sidebarCollapsed }" @click="toggleSidebar">
          <span class="sidebar-toggle__icon">{{ sidebarCollapsed ? '›' : '‹' }}</span>
          <span v-if="!sidebarCollapsed" class="sr-only">{{ t.collapseSidebar }}</span>
          <span v-else class="sr-only">{{ t.expandSidebar }}</span>
        </button>
        <div v-if="!sidebarCollapsed" class="sidebar-header__title">
          <p class="eyebrow">MongoDB</p>
          <h2>{{ t.databaseView }}</h2>
        </div>
        <div v-if="!sidebarCollapsed" class="sidebar-header__actions">
          <el-button v-if="!sidebarCollapsed" text @click="loadTree">{{ t.refresh }}</el-button>
        </div>
      </div>

      <el-scrollbar v-if="!sidebarCollapsed">
        <div v-loading="treeLoading" class="sidebar-body">
          <div v-if="!treeLoading && databases.length === 0" class="sidebar-empty">
            <p>{{ t.noBusinessDatabases }}</p>
            <span>{{ t.noBusinessDatabasesHint }}</span>
            <el-button type="primary" plain size="small" @click="openCreateDatabaseDialog">{{ t.createFirstDatabase }}</el-button>
          </div>

          <div v-for="database in databases" :key="database" class="database-group">
            <div class="database-row">
              <button class="database-button" :class="{ active: selectedDatabase === database }" @click="toggleDatabase(database)">
                <span class="database-button__title">{{ database }}</span>
                <span v-if="collectionCount(database)" class="database-meta">{{ collectionCount(database) }}</span>
              </button>
            </div>

            <div v-if="collectionsByDatabase[database]?.length" class="collection-list">
              <button
                v-for="collection in collectionsByDatabase[database]"
                :key="`${database}-${collection.name}`"
                class="collection-button"
                :class="{ active: selectedDatabase === database && selectedCollection === collection.name }"
                @click="selectCollection(database, collection.name)"
              >
                <div class="collection-main">
                  <span class="collection-main__title">{{ collection.name }}</span>
                  <span v-if="collection.is_system" class="collection-tag collection-tag--system">{{ t.systemTag }}</span>
                </div>
                <span v-if="typeof collection.document_count === 'number'" class="collection-count">
                  {{ t.documentCountSuffix(collection.document_count) }}
                </span>
              </button>
            </div>
          </div>
        </div>
      </el-scrollbar>
    </aside>

    <main class="console-main">
      <header class="topbar">
        <div class="cluster-info">
          <span>{{ activeConnectionLabel }}</span>
          <span class="divider">/</span>
          <span>{{ sessionConnection?.database || connection?.database || 'admin' }}</span>
        </div>
        <div class="status-area">
          <el-dropdown trigger="click" @command="handleLanguageChange">
            <el-button text>{{ t.language }} / {{ locale.toUpperCase() }}</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item v-for="language in languages" :key="language.value" :command="language.value">
                  {{ language.label }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <div class="status-pill">
            <span class="status-dot" :class="{ healthy: connection?.status === 'ok', error: connection?.status !== 'ok' }"></span>
            <span>{{ connection?.status === 'ok' ? t.connected : t.unavailable }}</span>
          </div>
          <el-button text @click="goToConnect">{{ t.changeConnection }}</el-button>
          <el-button text @click="disconnectSession">{{ t.disconnect }}</el-button>
          <el-button text @click="retryConnection">{{ t.retry }}</el-button>
        </div>
      </header>

      <section class="hero-card">
        <div class="hero-card__content">
          <p class="eyebrow">{{ t.mongoConsole }}</p>
          <h1>{{ selectedCollection || t.heroTitle }}</h1>
          <p class="hero-summary">{{ t.heroSummary }}</p>

          <div class="hero-metrics">
            <div class="metric-card">
              <span>{{ t.databaseCount }}</span>
              <strong>{{ databases.length }}</strong>
            </div>
            <div class="metric-card">
              <span>{{ t.visibleDocuments }}</span>
              <strong>{{ documents.length }}</strong>
            </div>
            <div class="metric-card">
              <span>{{ t.roundTrip }}</span>
              <strong>{{ connection?.round_trip_ms ?? '-' }} ms</strong>
            </div>
          </div>
        </div>

      </section>

      <section class="stats-strip">
        <div class="strip-card">
          <span>{{ t.currentDatabase }}</span>
          <strong>{{ connection?.database || '-' }}</strong>
        </div>
        <div class="strip-card">
          <span>{{ t.version }}</span>
          <strong>{{ connection?.version || '-' }}</strong>
        </div>
        <div class="strip-card">
          <span>{{ t.currentSelection }}</span>
          <strong>{{ collectionTitle }}</strong>
        </div>
        <div class="strip-card">
          <span>{{ t.totalHits }}</span>
          <strong>{{ total }}</strong>
        </div>
      </section>

      <section class="workspace-card">
        <div class="workspace-header">
          <div>
            <p class="eyebrow">{{ t.queryPanel }}</p>
            <h2>{{ t.documentStream }}</h2>
          </div>
        </div>

        <div class="control-panels">
          <section class="control-panel">
            <div class="control-panel__header">
              <p class="eyebrow">{{ t.database }}</p>
              <h3>{{ t.databaseActions }}</h3>
            </div>
            <div class="control-panel__actions">
              <el-button type="primary" plain @click="openCreateDatabaseDialog">{{ t.createDatabase }}</el-button>
              <el-button type="danger" plain :disabled="!selectedDatabase" @click="confirmDeleteDatabase(selectedDatabase)">{{ t.deleteDatabaseAction }}</el-button>
            </div>
          </section>

          <section class="control-panel">
            <div class="control-panel__header">
              <p class="eyebrow">{{ t.collection }}</p>
              <h3>{{ t.collectionActions }}</h3>
            </div>
            <div class="control-panel__actions">
              <el-button plain :disabled="!selectedDatabase" @click="openCreateCollection(selectedDatabase)">{{ t.newCollection }}</el-button>
              <el-button type="danger" plain :disabled="!selectedCollection" @click="confirmDeleteCollection">{{ t.deleteCollection }}</el-button>
              <el-button plain :disabled="!selectedCollection" @click="handleBackupCommand('json')">{{ t.backupCollection }}</el-button>
              <el-button plain :disabled="!selectedDatabase" @click="handleBackupCommand('restore')">{{ t.restoreFromBackup }}</el-button>
              <el-button plain :disabled="!selectedCollection" @click="loadIndexes">{{ t.refreshIndexes }}</el-button>
              <el-button plain :disabled="!selectedCollection" @click="openIndexDialog">{{ t.createIndex }}</el-button>
            </div>

            <div class="index-panel">
              <div class="index-panel__header">
                <strong>{{ t.indexPanel }}</strong>
                <span v-if="selectedCollection">{{ indexes.length }}</span>
              </div>
              <div v-if="!selectedCollection" class="index-panel__empty">{{ t.noIndexes }}</div>
              <div v-else-if="indexesLoading" class="index-panel__empty">{{ t.refresh }}</div>
              <div v-else-if="indexes.length === 0" class="index-panel__empty">{{ t.noIndexes }}</div>
              <div v-else class="index-list">
                <div v-for="index in indexes" :key="index.name" class="index-item">
                  <div class="index-item__main">
                    <strong>{{ index.name }}</strong>
                    <code>{{ formatIndexKeys(index.keys) }}</code>
                    <div class="index-item__flags">
                      <span v-if="index.unique">unique</span>
                      <span v-if="index.sparse">sparse</span>
                    </div>
                  </div>
                  <el-button text type="danger" size="small" :disabled="index.name === '_id_'" @click="confirmDeleteIndex(index.name)">
                    {{ t.delete }}
                  </el-button>
                </div>
              </div>
            </div>
          </section>
        </div>

        <div class="query-panel">
          <div class="query-panel__toolbar">
            <el-segmented :model-value="queryMode" :options="queryModeOptions" @change="handleQueryModeChange" />
            <el-button :disabled="!selectedCollection" @click="loadDocuments">{{ t.runQuery }}</el-button>
            <el-button plain :disabled="!selectedCollection" @click="resetQuery">{{ t.resetQuery }}</el-button>
          </div>
          <template v-if="queryMode === 'builder'">
            <div class="query-panel__header">
              <el-button plain size="small" @click="addConditionRow">{{ t.addCondition }}</el-button>
            </div>

            <div v-for="(condition, index) in conditions" :key="index" class="condition-row">
              <el-input v-model="condition.field" :placeholder="t.field" class="condition-field" />
              <el-select v-model="condition.operator" class="condition-operator">
                <el-option v-for="item in operatorOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
              <el-select v-model="condition.valueType" class="condition-type">
                <el-option v-for="item in valueTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
              </el-select>
              <el-input v-model="condition.value" :placeholder="t.value" class="condition-value" />
              <el-select v-if="index < conditions.length - 1" v-model="condition.join" class="condition-join">
                <el-option :label="t.and" value="and" />
                <el-option :label="t.or" value="or" />
              </el-select>
              <span v-else class="condition-join condition-join--empty"></span>
              <el-button text @click="removeConditionRow(index)">{{ t.remove }}</el-button>
            </div>
          </template>

          <template v-else>
            <el-input v-model="filterInput" type="textarea" :rows="4" :placeholder="t.invalidFilter" />
          </template>
        </div>

        <div class="workspace-body">
          <section class="table-card">
            <div class="table-header">
              <div>
                <p class="eyebrow">{{ t.collection }}</p>
                <h2>{{ collectionTitle }}</h2>
              </div>
              <div class="table-header__right">
                <div class="table-meta">
                  <span>{{ t.pageSize(pageSize) }}</span>
                  <span>{{ t.totalDocuments(total) }}</span>
                </div>
                <div class="table-actions">
                  <el-button type="primary" :disabled="!selectedCollection" @click="openCreateDocument">{{ t.newDocument }}</el-button>
                  <el-dropdown :disabled="!selectedCollection" @command="handleImportCommand">
                    <el-button plain :disabled="!selectedCollection">{{ t.importDocument }}</el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="json">{{ t.importJson }}</el-dropdown-item>
                        <el-dropdown-item command="csv">{{ t.importCsv }}</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                  <el-dropdown :disabled="!selectedCollection" @command="handleExportCommand">
                    <el-button plain :disabled="!selectedCollection">{{ t.exportDocument }}</el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="json">{{ t.exportJson }}</el-dropdown-item>
                        <el-dropdown-item command="csv">{{ t.exportCsv }}</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                  <el-dropdown :disabled="!selectedCollection || selectedCount === 0" @command="handleSelectedExportCommand">
                    <el-button plain :disabled="!selectedCollection || selectedCount === 0">{{ t.exportSelected }}</el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="json">{{ t.exportJson }}</el-dropdown-item>
                        <el-dropdown-item command="csv">{{ t.exportCsv }}</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                  <el-button type="danger" plain :disabled="!selectedCollection || selectedCount === 0" @click="confirmBulkDelete">{{ t.bulkDelete }}</el-button>
                </div>
              </div>
            </div>

            <div v-if="!treeLoading && databases.length === 0" class="empty-state">
              <p>{{ t.noBusinessDatabases }}</p>
              <span>{{ t.noBusinessDatabasesHint }}</span>
              <div class="empty-actions">
                <el-button type="primary" @click="openCreateDatabaseDialog">{{ t.createFirstDatabase }}</el-button>
              </div>
            </div>

            <div v-else-if="!selectedCollection" class="empty-state">
              <p>{{ databases.length ? t.selectCollection : t.noBusinessDatabases }}</p>
              <span>{{ databases.length ? t.selectCollectionHint : t.selectDatabaseToManage }}</span>
            </div>

            <template v-else-if="databases.length">
              <div v-if="documentsError" class="empty-state error">
                <p>{{ t.queryFailed }}</p>
                <span>{{ documentsError }}</span>
              </div>

            <div v-else-if="!documentsLoading && documents.length === 0" class="empty-state">
              <p>{{ t.emptyCollection }}</p>
              <span>{{ t.emptyCollectionHint }}</span>
              <div class="empty-actions">
                  <el-button type="primary" @click="openCreateDocument">{{ t.createFirstDocument }}</el-button>
                  <el-button plain @click="loadDocuments">{{ t.refresh }}</el-button>
                </div>
              </div>

              <div v-else class="documents-scroll">
                <el-table
                  ref="documentsTable"
                  :data="documents"
                  v-loading="documentsLoading"
                  table-layout="auto"
                  @selection-change="handleSelectionChange"
                  @row-click="handleDocumentRowClick"
                >
                  <el-table-column type="selection" width="52" />
                  <el-table-column label="_id" min-width="220">
                    <template #default="{ row }">
                      <code class="doc-id">{{ formatCell(row._id) }}</code>
                    </template>
                  </el-table-column>
                  <el-table-column :label="t.document" min-width="560">
                    <template #default="{ row }">
                      <pre class="json-preview">{{ formatDocument(row) }}</pre>
                    </template>
                  </el-table-column>
                  <el-table-column :label="t.action" width="180" fixed="right">
                    <template #default="{ row }">
                      <div class="row-actions">
                        <el-button link type="primary" @click="openEditDocument(row)">{{ t.edit }}</el-button>
                        <el-button link type="danger" @click="confirmDeleteDocument(row)">{{ t.delete }}</el-button>
                      </div>
                    </template>
                  </el-table-column>
                </el-table>
              </div>

              <div v-if="selectedCount > 0" class="selection-summary">
                {{ t.selectedCount(selectedCount) }}
              </div>

              <div class="pagination-row">
                <el-pagination
                  layout="total, prev, pager, next"
                  :total="total"
                  :current-page="page"
                  :page-size="pageSize"
                  @current-change="handlePageChange"
                />
              </div>
            </template>
          </section>

        </div>
      </section>
    </main>

    <input ref="importInput" type="file" class="hidden-input" @change="handleImportFileChange" />

    <el-dialog v-model="databaseDialogVisible" :title="t.createDatabaseTitle" width="480px">
      <el-form label-width="130px">
        <el-form-item :label="t.database">
          <el-input v-model="databaseForm.database" />
        </el-form-item>
        <el-form-item :label="t.firstCollection">
          <el-input v-model="databaseForm.firstCollection" />
        </el-form-item>
      </el-form>
      <p class="dialog-hint">{{ t.createDatabasePrompt }}</p>
      <template #footer>
        <el-button @click="databaseDialogVisible = false">{{ t.cancel }}</el-button>
        <el-button type="primary" @click="submitCreateDatabase">{{ t.confirm }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="indexDialogVisible" :title="t.createIndexTitle" width="480px">
      <el-form label-width="100px">
        <el-form-item :label="t.indexName">
          <el-input v-model="indexForm.name" />
        </el-form-item>
        <el-form-item :label="t.indexField">
          <el-input v-model="indexForm.field" />
        </el-form-item>
        <el-form-item :label="t.indexOrder">
          <el-select v-model="indexForm.order">
            <el-option label="ASC" :value="1" />
            <el-option label="DESC" :value="-1" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t.uniqueIndex">
          <el-switch v-model="indexForm.unique" />
        </el-form-item>
        <el-form-item :label="t.sparseIndex">
          <el-switch v-model="indexForm.sparse" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="indexDialogVisible = false">{{ t.cancel }}</el-button>
        <el-button type="primary" @click="submitCreateIndex">{{ t.confirm }}</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="editorVisible" :title="editorTitle" size="52%" class="editor-drawer">
      <div class="editor-meta">
        <p class="eyebrow">{{ t.editorTitle }}</p>
        <p class="editor-hint">{{ t.extendedJsonHint }}</p>
      </div>

      <div class="editor-toolbar">
        <el-segmented :model-value="editorMode" :options="editorModeOptions" @change="handleEditorModeChange" />
      </div>

      <div v-if="editorMode === 'form'" class="form-editor">
        <div v-if="complexFields.length" class="complex-fields-tip">
          {{ t.complexFieldsHint(complexFields) }}
        </div>

        <div class="form-editor__header">
          <span>{{ t.simpleFields }}</span>
          <el-button plain size="small" @click="addFormField">{{ t.addField }}</el-button>
        </div>

        <div v-for="(field, index) in formFields" :key="index" class="form-row">
          <el-input v-model="field.key" :placeholder="t.fieldName" class="form-key" />
          <el-select v-model="field.type" class="form-type">
            <el-option v-for="item in valueTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-input v-model="field.value" :placeholder="t.value" class="form-value" />
          <el-button text @click="removeFormField(index)">{{ t.remove }}</el-button>
        </div>
      </div>

      <div v-else class="editor-surface">
        <el-input v-model="editorValue" type="textarea" :rows="24" placeholder="{\n  \n}" />
      </div>

      <template #footer>
        <div class="drawer-footer">
          <el-button @click="editorVisible = false">{{ t.cancel }}</el-button>
          <el-button type="primary" @click="saveDocument">{{ t.save }}</el-button>
        </div>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { ElMessage, ElMessageBox, ElTable } from 'element-plus'
import { useRouter } from 'vue-router'

import {
  backupCollection,
  bulkDeleteDocuments,
  createCollection,
  createDatabase,
  createDocument,
  createIndex,
  deleteCollection,
  deleteDatabase,
  deleteDocument,
  deleteIndex,
  exportDocuments,
  fetchCollections,
  fetchConnectionInfo,
  fetchDatabases,
  fetchDocument,
  fetchDocuments,
  fetchIndexes,
  importDocuments,
  restoreCollection,
  updateDocument,
} from '../api/mongo'
import type { CollectionItem, ConnectionInfo, IndexItem } from '../api/types'
import { useMongoI18n } from '../i18n/mongodb'
import { clearMongoSessionConnection, getMongoSessionConnection } from '../session/mongoConnection'

const router = useRouter()
const { locale, setLocale, t, languages } = useMongoI18n()

type ConnectionState = Partial<ConnectionInfo> & { status: string }
const connection = ref<ConnectionState | null>(null)
const databases = ref<string[]>([])
const collectionsByDatabase = ref<Record<string, CollectionItem[]>>({})
const treeLoading = ref(false)
const sidebarCollapsed = ref(false)
const sessionConnection = ref(getMongoSessionConnection())

const selectedDatabase = ref('')
const selectedCollection = ref('')
const documents = ref<Record<string, unknown>[]>([])
const documentsLoading = ref(false)
const documentsError = ref('')
const documentsTable = ref<InstanceType<typeof ElTable> | null>(null)
const page = ref(1)
const pageSize = 20
const total = ref(0)
const selectedRows = ref<Record<string, unknown>[]>([])

const queryMode = ref<'builder' | 'json'>('builder')
const builderLogic = ref<'and' | 'or'>('and')
const filterInput = ref('')
const conditions = ref([createConditionRow()])

const editorVisible = ref(false)
const editorTitle = ref('document')
const editorMode = ref<'form' | 'json'>('form')
const editorValue = ref('{\n  \n}')
const editingId = ref('')
const originalDocument = ref<Record<string, unknown>>({})
const formFields = ref([createFormField()])
const originalSimpleKeys = ref<string[]>([])
const complexFields = ref<string[]>([])

const importInput = ref<HTMLInputElement | null>(null)
const pendingImportFormat = ref<'json' | 'csv'>('json')
const importAction = ref<'query' | 'restore'>('query')
const pendingRestoreCollection = ref('')
const databaseDialogVisible = ref(false)
const databaseForm = ref({
  database: '',
  firstCollection: '',
})
const indexes = ref<IndexItem[]>([])
const indexesLoading = ref(false)
const indexDialogVisible = ref(false)
const indexForm = ref({
  name: '',
  field: '',
  order: 1,
  unique: false,
  sparse: false,
})

const queryModeOptions = computed(() => [
  { label: t.value.builderMode, value: 'builder' },
  { label: t.value.jsonMode, value: 'json' },
])

const editorModeOptions = computed(() => [
  { label: t.value.formMode, value: 'form' },
  { label: t.value.rawJsonMode, value: 'json' },
])

const operatorOptions = computed(() => [
  { label: '=', value: 'eq' },
  { label: '!=', value: 'ne' },
  { label: '>', value: 'gt' },
  { label: '>=', value: 'gte' },
  { label: '<', value: 'lt' },
  { label: '<=', value: 'lte' },
  { label: 'contains', value: 'contains' },
  { label: 'in', value: 'in' },
])

const valueTypeOptions = computed(() => [
  { label: 'string', value: 'string' },
  { label: 'number', value: 'number' },
  { label: 'boolean', value: 'boolean' },
  { label: 'date', value: 'date' },
  { label: 'objectId', value: 'objectid' },
  { label: 'null', value: 'null' },
])

const collectionTitle = computed(() =>
  selectedCollection.value ? `${selectedDatabase.value} / ${selectedCollection.value}` : t.value.notSelected,
)

const activeConnectionLabel = computed(() => {
  if (sessionConnection.value) {
    return `${sessionConnection.value.host}:${sessionConnection.value.port}`
  }
  return connection.value?.server || 'MongoDB'
})

const selectedCount = computed(() => selectedRows.value.length)

void bootstrap()

async function bootstrap() {
  if (!sessionConnection.value) {
    await router.replace({ name: 'connect', query: { redirect: '/' } })
    return
  }
  await Promise.all([loadConnection(), loadTree()])
}

async function loadConnection() {
  try {
    connection.value = await fetchConnectionInfo()
  } catch (error) {
    connection.value = { status: 'error' }
    ElMessage.error((error as Error).message)
  }
}

async function retryConnection() {
  sessionConnection.value = getMongoSessionConnection()
  await Promise.all([loadConnection(), loadTree()])
}

async function goToConnect() {
  await router.push({ name: 'connect', query: { redirect: '/' } })
}

async function disconnectSession() {
  clearMongoSessionConnection()
  sessionConnection.value = null
  await router.replace({ name: 'connect', query: { redirect: '/' } })
}

function handleLanguageChange(value: string | number | boolean) {
  setLocale(String(value))
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

async function loadTree() {
  treeLoading.value = true
  try {
    const data = await fetchDatabases()
    databases.value = data.items || []
    if (!databases.value.length) {
      collectionsByDatabase.value = {}
      selectedDatabase.value = ''
      selectedCollection.value = ''
      documents.value = []
      total.value = 0
      documentsError.value = ''
      indexes.value = []
      return
    }

    if (selectedDatabase.value && !databases.value.includes(selectedDatabase.value)) {
      selectedDatabase.value = ''
      selectedCollection.value = ''
      documents.value = []
      total.value = 0
      documentsError.value = ''
      indexes.value = []
    }
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    treeLoading.value = false
  }
}

function openCreateDatabaseDialog() {
  databaseForm.value = {
    database: '',
    firstCollection: '',
  }
  databaseDialogVisible.value = true
}

async function submitCreateDatabase() {
  const database = databaseForm.value.database.trim()
  const firstCollection = databaseForm.value.firstCollection.trim()

  if (!database || !firstCollection) {
    ElMessage.error(t.value.createDatabasePrompt)
    return
  }

  try {
    await createDatabase({
      database,
      firstCollection,
    })
    databaseDialogVisible.value = false
    await loadTree()
    await toggleDatabase(database, true)
    selectedDatabase.value = database
    selectedCollection.value = firstCollection
    ElMessage.success(t.value.databaseCreated)
    await Promise.all([loadDocuments(), loadIndexes()])
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

function collectionCount(database: string) {
  if (!collectionsByDatabase.value[database]) {
    return ''
  }
  return `${collectionsByDatabase.value[database].length} ${t.value.collection}`
}

async function toggleDatabase(database: string, forceRefresh = false) {
  selectedDatabase.value = database
  if (collectionsByDatabase.value[database] && !forceRefresh) {
    const next = { ...collectionsByDatabase.value }
    delete next[database]
    collectionsByDatabase.value = next
    if (selectedCollection.value && selectedDatabase.value === database) {
      selectedCollection.value = ''
    }
    return
  }

  try {
    const data = await fetchCollections(database, { includeCounts: true })
    collectionsByDatabase.value = {
      ...collectionsByDatabase.value,
      [database]: data.items || [],
    }
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

function selectCollection(database: string, collection: string) {
  selectedDatabase.value = database
  selectedCollection.value = collection
  page.value = 1
  documents.value = []
  total.value = 0
  documentsError.value = ''
  void Promise.all([loadDocuments(), loadIndexes()])
}

function buildQueryParams() {
  const params: Record<string, unknown> = {
    page: page.value,
    pageSize,
  }

  if (queryMode.value === 'json') {
    if (filterInput.value.trim()) {
      try {
        JSON.parse(filterInput.value)
      } catch {
        throw new Error(t.value.invalidFilter)
      }
      params.filter = filterInput.value
    }
    return params
  }

  const activeConditions = conditions.value.filter((item) => item.field && item.operator)
  if (!activeConditions.length) {
    return params
  }

  const hasInvalid = activeConditions.some((item) => item.valueType !== 'null' && String(item.value ?? '').trim() === '')
  if (hasInvalid) {
    throw new Error(t.value.invalidBuilder)
  }

  params.logic = builderLogic.value
  params.conditions = JSON.stringify(
    activeConditions.map((item) => ({
      field: item.field,
      operator: item.operator,
      value_type: item.valueType,
      value: normalizeConditionValue(item),
      join: item.join,
    })),
  )
  return params
}

async function loadDocuments() {
  if (!selectedDatabase.value || !selectedCollection.value) return

  let params: Record<string, unknown>
  try {
    params = buildQueryParams()
  } catch (error) {
    documentsError.value = (error as Error).message
    ElMessage.error(documentsError.value)
    return
  }

  documentsLoading.value = true
  documentsError.value = ''
  try {
    const data = await fetchDocuments(selectedDatabase.value, selectedCollection.value, params as { page: number; pageSize: number; filter?: string; conditions?: string; logic?: string })
    documents.value = data.items || []
    total.value = data.total || 0
    selectedRows.value = []
  } catch (error) {
    documents.value = []
    total.value = 0
    documentsError.value = (error as Error).message
    ElMessage.error(documentsError.value)
  } finally {
    documentsLoading.value = false
  }
}

function handleQueryModeChange(value: string | number | boolean) {
  queryMode.value = value as 'builder' | 'json'
}

function addConditionRow() {
  conditions.value.push(createConditionRow())
}

function removeConditionRow(index: number) {
  conditions.value.splice(index, 1)
  if (!conditions.value.length) {
    conditions.value.push(createConditionRow())
  }
}

function resetQuery() {
  filterInput.value = ''
  builderLogic.value = 'and'
  conditions.value = [createConditionRow()]
  page.value = 1
  documents.value = []
  total.value = 0
  documentsError.value = ''
  if (selectedCollection.value) {
    void loadDocuments()
  }
}

async function openCreateCollection(database: string) {
  if (!database) {
    ElMessage.warning(t.value.selectDatabaseFirst)
    return
  }

  try {
    const { value } = await ElMessageBox.prompt(t.value.createCollectionPrompt(database), t.value.createCollectionTitle, {
      confirmButtonText: t.value.confirm,
      cancelButtonText: t.value.cancel,
    })
    await createCollection(database, value)
    await toggleDatabase(database, true)
    selectedDatabase.value = database
    selectedCollection.value = value
    documents.value = []
    total.value = 0
    documentsError.value = ''
    ElMessage.success(t.value.collectionCreated)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error((error as Error).message)
    }
  }
}

function openCreateDocument() {
  editingId.value = ''
  originalDocument.value = {}
  editorTitle.value = t.value.newDocument
  editorMode.value = 'form'
  formFields.value = [createFormField()]
  originalSimpleKeys.value = []
  complexFields.value = []
  editorValue.value = '{\n  \n}'
  editorVisible.value = true
}

async function openEditDocument(row: Record<string, unknown>) {
  const id = extractObjectId(row._id)
  if (!id) {
    ElMessage.error(t.value.invalidObjectId)
    return
  }

  try {
    const data = await fetchDocument(selectedDatabase.value, selectedCollection.value, id)
    editingId.value = id
    editorTitle.value = `${t.value.document} ${id}`
    originalDocument.value = deepClone(data.item || {})
    editorValue.value = JSON.stringify(data.item, null, 2)
    buildFormFieldsFromDocument(data.item || {})
    editorVisible.value = true
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

function handleEditorModeChange(value: string | number | boolean) {
  editorMode.value = value as 'form' | 'json'
  if (editorMode.value === 'json') {
    try {
      editorValue.value = JSON.stringify(buildDocumentFromForm(), null, 2)
    } catch {
      editorValue.value = JSON.stringify(originalDocument.value, null, 2)
    }
  } else {
    let source = originalDocument.value
    try {
      source = JSON.parse(editorValue.value)
    } catch {
      source = originalDocument.value
    }
    buildFormFieldsFromDocument(source)
  }
}

function addFormField() {
  formFields.value.push(createFormField())
}

function removeFormField(index: number) {
  formFields.value.splice(index, 1)
  if (!formFields.value.length) {
    formFields.value.push(createFormField())
  }
}

async function saveDocument() {
  if (!selectedDatabase.value || !selectedCollection.value) return

  let value: Record<string, unknown>
  try {
    value = editorMode.value === 'form' ? buildDocumentFromForm() : JSON.parse(editorValue.value)
  } catch (error) {
    ElMessage.error((error as Error).message || t.value.invalidJson)
    return
  }

  try {
    if (editingId.value) {
      await updateDocument(selectedDatabase.value, selectedCollection.value, editingId.value, value)
      ElMessage.success(t.value.documentUpdated)
    } else {
      await createDocument(selectedDatabase.value, selectedCollection.value, value)
      ElMessage.success(t.value.documentCreated)
    }
    editorVisible.value = false
    await loadDocuments()
    await toggleDatabase(selectedDatabase.value, true)
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function confirmDeleteDocument(row: Record<string, unknown>) {
  const id = extractObjectId(row._id)
  if (!id) {
    ElMessage.error(t.value.invalidObjectId)
    return
  }

  try {
    await ElMessageBox.confirm(t.value.deleteDocumentConfirm(id), t.value.deleteDocumentTitle, {
      type: 'warning',
      confirmButtonText: t.value.delete,
      cancelButtonText: t.value.cancel,
    })
    await deleteDocument(selectedDatabase.value, selectedCollection.value, id)
    ElMessage.success(t.value.documentDeleted)
    await loadDocuments()
    await toggleDatabase(selectedDatabase.value, true)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error((error as Error).message)
    }
  }
}

async function confirmDeleteCollection() {
  if (!selectedDatabase.value || !selectedCollection.value) return

  try {
    await ElMessageBox.confirm(
      t.value.deleteCollectionConfirm(selectedDatabase.value, selectedCollection.value),
      t.value.deleteCollectionTitle,
      {
        type: 'warning',
        confirmButtonText: t.value.delete,
        cancelButtonText: t.value.cancel,
      },
    )
    await deleteCollection(selectedDatabase.value, selectedCollection.value)
    await toggleDatabase(selectedDatabase.value, true)
    documents.value = []
    total.value = 0
    documentsError.value = ''
    selectedCollection.value = ''
    ElMessage.success(t.value.collectionDeleted)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error((error as Error).message)
    }
  }
}

async function confirmDeleteDatabase(database: string) {
  try {
    await ElMessageBox.confirm(
      t.value.deleteDatabaseConfirm(database),
      t.value.deleteDatabaseTitle,
      {
        type: 'warning',
        confirmButtonText: t.value.delete,
        cancelButtonText: t.value.cancel,
      },
    )
    await deleteDatabase(database)
    if (selectedDatabase.value === database) {
      selectedDatabase.value = ''
      selectedCollection.value = ''
      documents.value = []
      total.value = 0
      documentsError.value = ''
      indexes.value = []
    }
    await loadTree()
    ElMessage.success(t.value.databaseDeleted)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error((error as Error).message)
    }
  }
}

function handlePageChange(nextPage: number) {
  page.value = nextPage
  void loadDocuments()
}

function handleSelectionChange(rows: Record<string, unknown>[]) {
  selectedRows.value = rows
}

function handleDocumentRowClick(row: Record<string, unknown>, _column: unknown, event: Event) {
  const target = event.target as HTMLElement | null
  if (target?.closest('.row-actions, .el-button, .el-checkbox')) {
    return
  }
  documentsTable.value?.toggleRowSelection(row)
}

async function handleExportCommand(format: string) {
  if (!selectedDatabase.value || !selectedCollection.value) return

  try {
    const params = buildQueryParams()
    params.format = format
    const response = await exportDocuments(selectedDatabase.value, selectedCollection.value, params)
    downloadBlob(response.data, response.headers['content-type'], response.headers['content-disposition'], `${selectedCollection.value}.${format}`)
  } catch (error) {
    ElMessage.error((error as Error).message || t.value.exportFailed)
  }
}

async function handleSelectedExportCommand(format: string) {
  if (!selectedDatabase.value || !selectedCollection.value || selectedRows.value.length === 0) return

  try {
    const ids = selectedRows.value
      .map((row) => extractObjectId(row._id))
      .filter((id): id is string => Boolean(id))

    const response = await exportDocuments(selectedDatabase.value, selectedCollection.value, {
      format,
      ids,
    })
    downloadBlob(response.data, response.headers['content-type'], response.headers['content-disposition'], `${selectedCollection.value}.selected.${format}`)
  } catch (error) {
    ElMessage.error((error as Error).message || t.value.exportFailed)
  }
}

async function handleBackupCommand(command: string) {
  if (!selectedDatabase.value) return

  if (command === 'restore') {
    try {
      const collection = await resolveRestoreCollection()
      if (!collection) return

      await ElMessageBox.confirm(t.value.restoreConfirm(collection), t.value.restoreTitle, {
        type: 'warning',
        confirmButtonText: t.value.confirm,
        cancelButtonText: t.value.cancel,
      })
      importAction.value = 'restore'
      pendingRestoreCollection.value = collection
      pendingImportFormat.value = 'json'
      if (importInput.value) {
        importInput.value.value = ''
        importInput.value.accept = '.json,application/json'
        importInput.value.click()
      }
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error((error as Error).message || t.value.restoreFailed)
      }
    }
    return
  }

  if (!selectedCollection.value) return

  try {
    const response = await backupCollection(selectedDatabase.value, selectedCollection.value)
    downloadBlob(response.data, response.headers['content-type'], response.headers['content-disposition'], `${selectedCollection.value}.backup.json`)
    ElMessage.success(t.value.backupSuccess)
  } catch (error) {
    ElMessage.error((error as Error).message || t.value.backupFailed)
  }
}

function handleImportCommand(format: string) {
  if (!selectedDatabase.value || !selectedCollection.value) return

  importAction.value = 'query'
  pendingRestoreCollection.value = ''
  pendingImportFormat.value = format as 'json' | 'csv'
  if (importInput.value) {
    importInput.value.value = ''
    importInput.value.accept = format === 'csv' ? '.csv,text/csv' : '.json,application/json'
    importInput.value.click()
  }
}

async function handleImportFileChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (!file || !selectedDatabase.value) return

  try {
    if (importAction.value === 'restore') {
      const collection = pendingRestoreCollection.value || selectedCollection.value
      if (!collection) return

      const res = await restoreCollection(selectedDatabase.value, collection, file)
      ElMessage.success(t.value.restoreResult((res.data as { restored_count?: number }).restored_count || 0))
      selectedCollection.value = collection
    } else {
      if (!selectedCollection.value) return

      const res = await importDocuments(selectedDatabase.value, selectedCollection.value, file, pendingImportFormat.value)
      ElMessage.success(t.value.importResult((res.data as { inserted_count?: number }).inserted_count || 0))
    }
    await loadDocuments()
    await toggleDatabase(selectedDatabase.value, true)
  } catch (error) {
    ElMessage.error((error as Error).message || (importAction.value === 'restore' ? t.value.restoreFailed : t.value.importFailed))
  }
}

async function resolveRestoreCollection() {
  if (selectedCollection.value) {
    return selectedCollection.value
  }

  try {
    const { value } = await ElMessageBox.prompt(t.value.restoreCollectionPrompt(selectedDatabase.value), t.value.restoreTitle, {
      confirmButtonText: t.value.confirm,
      cancelButtonText: t.value.cancel,
      inputValidator: (input) => Boolean(String(input || '').trim()) || t.value.restoreCollectionRequired,
    })
    return String(value || '').trim()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error((error as Error).message)
    }
    return ''
  }
}

async function confirmBulkDelete() {
  if (!selectedRows.value.length || !selectedCollection.value) return

  try {
    await ElMessageBox.confirm(t.value.bulkDeleteConfirm(selectedRows.value.length), t.value.bulkDeleteTitle, {
      type: 'warning',
      confirmButtonText: t.value.delete,
      cancelButtonText: t.value.cancel,
    })

    const ids = selectedRows.value
      .map((row) => extractObjectId(row._id))
      .filter((id): id is string => Boolean(id))

    const result = await bulkDeleteDocuments(selectedDatabase.value, selectedCollection.value, ids)
    ElMessage.success(t.value.bulkDeleteSuccess((result.deleted_count as number) || 0))
    await loadDocuments()
    await toggleDatabase(selectedDatabase.value, true)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error((error as Error).message)
    }
  }
}

async function loadIndexes() {
  if (!selectedDatabase.value || !selectedCollection.value) {
    indexes.value = []
    return
  }

  indexesLoading.value = true
  try {
    const data = await fetchIndexes(selectedDatabase.value, selectedCollection.value)
    indexes.value = data.items || []
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    indexesLoading.value = false
  }
}

function openIndexDialog() {
  indexForm.value = {
    name: '',
    field: '',
    order: 1,
    unique: false,
    sparse: false,
  }
  indexDialogVisible.value = true
}

async function submitCreateIndex() {
  if (!selectedDatabase.value || !selectedCollection.value) return

  try {
    await createIndex(selectedDatabase.value, selectedCollection.value, indexForm.value)
    indexDialogVisible.value = false
    ElMessage.success(t.value.createIndexSuccess)
    await loadIndexes()
  } catch (error) {
    ElMessage.error((error as Error).message)
  }
}

async function confirmDeleteIndex(name: string) {
  try {
    await ElMessageBox.confirm(t.value.deleteIndexConfirm(name), t.value.indexPanel, {
      type: 'warning',
      confirmButtonText: t.value.delete,
      cancelButtonText: t.value.cancel,
    })
    await deleteIndex(selectedDatabase.value, selectedCollection.value, name)
    ElMessage.success(t.value.deleteIndexSuccess)
    await loadIndexes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error((error as Error).message)
    }
  }
}

function buildFormFieldsFromDocument(document: Record<string, unknown>) {
  const fields: { key: string; type: string; value: string }[] = []
  const complex: string[] = []
  const simpleKeys: string[] = []

  Object.entries(document || {}).forEach(([key, value]) => {
    if (key === '_id') return
    if (isSimpleEditableValue(value)) {
      fields.push({
        key,
        type: inferValueType(value),
        value: valueToFormString(value),
      })
      simpleKeys.push(key)
    } else {
      complex.push(key)
    }
  })

  formFields.value = fields.length ? fields : [createFormField()]
  originalSimpleKeys.value = simpleKeys
  complexFields.value = complex
  editorMode.value = complex.length ? 'json' : 'form'
}

function buildDocumentFromForm() {
  const base = deepClone(originalDocument.value)
  const next: Record<string, unknown> = { ...base }

  originalSimpleKeys.value.forEach((key) => {
    delete next[key]
  })

  formFields.value.forEach((field) => {
    if (!field.key) return
    next[field.key] = parseFormValue(field)
  })

  return next
}

function formatDocument(row: Record<string, unknown>) {
  return JSON.stringify(row, null, 2)
}

function formatCell(value: unknown) {
  if (value && typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value ?? '')
}

function extractObjectId(value: unknown) {
  if (typeof value === 'string') return value
  if (value && typeof value === 'object' && '$oid' in (value as Record<string, unknown>)) {
    return String((value as Record<string, unknown>).$oid)
  }
  return null
}

function normalizeConditionValue(condition: { operator: string; valueType: string; value: string }) {
  if (condition.operator === 'in') {
    return String(condition.value || '')
      .split(',')
      .map((item) => item.trim())
      .filter(Boolean)
  }
  if (condition.valueType === 'number') {
    return Number(condition.value)
  }
  if (condition.valueType === 'boolean') {
    return String(condition.value).toLowerCase() === 'true'
  }
  if (condition.valueType === 'null') {
    return null
  }
  return condition.value
}

function createConditionRow() {
  return {
    field: '',
    operator: 'eq',
    valueType: 'string',
    value: '',
    join: 'and',
  }
}

function createFormField() {
  return {
    key: '',
    type: 'string',
    value: '',
  }
}

function isSimpleEditableValue(value: unknown) {
  return value === null || ['string', 'number', 'boolean'].includes(typeof value)
}

function inferValueType(value: unknown) {
  if (value === null) return 'null'
  if (typeof value === 'number') return 'number'
  if (typeof value === 'boolean') return 'boolean'
  return 'string'
}

function valueToFormString(value: unknown) {
  if (value === null) return ''
  return String(value)
}

function parseFormValue(field: { type: string; value: string }) {
  switch (field.type) {
    case 'number':
      return Number(field.value)
    case 'boolean':
      return String(field.value).toLowerCase() === 'true'
    case 'null':
      return null
    default:
      return field.value
  }
}

function deepClone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value ?? {})) as T
}

function downloadBlob(data: BlobPart, type: string | undefined, disposition: string | undefined, fallbackName: string) {
  const blob = new Blob([data], { type: type || 'application/octet-stream' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  const match = disposition?.match(/filename="([^"]+)"/)
  link.download = match ? match[1] : fallbackName
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.URL.revokeObjectURL(url)
}

function formatIndexKeys(keys: Record<string, unknown>) {
  return Object.entries(keys)
    .map(([key, value]) => `${key}:${value}`)
    .join(', ')
}
</script>

<style scoped>
.mongodb-console {
  display: grid;
  grid-template-columns: 320px 1fr;
  min-height: calc(100vh - 40px);
  gap: 20px;
}

.console-sidebar,
.hero-card,
.stats-strip,
.workspace-card {
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 6px 24px rgba(15, 23, 42, 0.06);
}

.console-sidebar {
  padding: 18px 16px;
  border: 1px solid #eef2f7;
  transition: width 0.25s ease, padding 0.25s ease;
  overflow: hidden;
}

.console-sidebar--collapsed {
  width: 76px;
  padding: 18px 16px;
}

.database-row,
.workspace-header,
.table-header,
.drawer-footer,
.pagination-row,
.row-actions,
.status-area,
.editor-toolbar,
.form-editor__header,
.query-panel__header {
  display: flex;
  align-items: center;
}

.database-row,
.workspace-header,
.table-header,
.form-editor__header,
.query-panel__header {
  justify-content: space-between;
  gap: 12px;
}

.sidebar-header {
  display: grid;
  grid-template-columns: 44px minmax(0, 1fr) auto;
  align-items: center;
  gap: 12px;
  min-height: 48px;
}

.sidebar-header--collapsed {
  grid-template-columns: 44px;
  justify-content: start;
}

.eyebrow {
  margin: 0 0 6px;
  color: #7c8aa5;
  font-size: 12px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.sidebar-body {
  padding: 12px 0 24px;
}

.database-group {
  margin-bottom: 16px;
}

.database-row__actions {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.database-button,
.collection-button {
  width: 100%;
  border: 0;
  background: transparent;
  text-align: left;
  cursor: pointer;
  border-radius: 12px;
  transition: background 0.2s ease, color 0.2s ease;
}

.database-button {
  display: flex;
  justify-content: space-between;
  padding: 12px;
  color: #1f2937;
}

.database-button.active,
.database-button:hover,
.collection-button.active,
.collection-button:hover {
  background: #edf4ff;
}

.collection-list {
  margin-top: 8px;
}

.collection-button {
  display: flex;
  justify-content: space-between;
  padding: 10px 12px 10px 24px;
  color: #4b5563;
}

.collection-main {
  display: flex;
  gap: 8px;
  align-items: center;
}

.collection-tag,
.database-meta,
.collection-count {
  color: #8b97ab;
  font-size: 12px;
}

.collection-tag--system {
  color: #c2410c;
}

.sidebar-empty {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 16px 8px 20px;
  color: #64748b;
}

.sidebar-empty p {
  margin: 0;
  color: #1f2937;
  font-weight: 600;
}

.console-main {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.cluster-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #4b5563;
}

.divider {
  color: #94a3b8;
}

.status-area {
  gap: 10px;
  flex-wrap: wrap;
}

.status-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  background: #fff;
  border-radius: 999px;
  padding: 8px 14px;
  box-shadow: 0 2px 10px rgba(15, 23, 42, 0.06);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 999px;
}

.status-dot.healthy {
  background: #22c55e;
}

.status-dot.error {
  background: #ef4444;
}

.hero-card {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  padding: 24px;
}

.hero-card h1 {
  margin: 0;
  font-size: 30px;
  line-height: 1.1;
}

.hero-summary {
  color: #4b5563;
  max-width: 720px;
}

.hero-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 18px;
}

.metric-card,
.strip-card {
  padding: 14px 16px;
  border-radius: 14px;
  background: #f8fafc;
}

.metric-card span,
.strip-card span {
  display: block;
  color: #64748b;
  font-size: 12px;
  margin-bottom: 4px;
}

.hero-actions {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  flex-wrap: wrap;
}

.stats-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  padding: 12px;
}

.workspace-card {
  padding: 20px;
}

.control-panels {
  display: grid;
  grid-template-columns: minmax(240px, 0.7fr) minmax(0, 1.3fr);
  gap: 14px;
  margin-top: 18px;
}

.control-panel {
  padding: 16px;
  border-radius: 16px;
  background: #f8fafc;
}

.control-panel__header h3 {
  margin: 0;
}

.control-panel__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.index-panel {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #e2e8f0;
}

.index-panel__header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 10px;
}

.index-panel__header span {
  color: #64748b;
  font-size: 12px;
}

.index-panel__empty {
  padding: 12px;
  border-radius: 12px;
  background: #fff;
  color: #64748b;
}

.toolbar-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.query-panel__toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.query-panel {
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  background: #f8fafc;
  border-radius: 16px;
  padding: 16px;
}

.condition-row {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr 0.9fr 1.1fr auto;
  gap: 10px;
  align-items: center;
}

.condition-join {
  min-width: 100px;
}

.condition-join--empty {
  min-width: 100px;
}

.workspace-body {
  display: block;
  width: 100%;
  margin-top: 18px;
}

.table-card {
  width: 100%;
  box-sizing: border-box;
  background: #f8fafc;
  border-radius: 16px;
  padding: 18px;
}

.table-header h2 {
  margin: 0;
}

.table-header__right {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-end;
}

.table-meta {
  display: flex;
  gap: 12px;
  color: #64748b;
  font-size: 13px;
}

.table-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  flex-wrap: wrap;
}

.empty-state {
  padding: 72px 20px;
  text-align: center;
  color: #64748b;
}

.empty-state p {
  margin: 0 0 8px;
  color: #1f2937;
  font-size: 16px;
}

.empty-state.error p {
  color: #dc2626;
}

.empty-actions {
  display: flex;
  justify-content: center;
  gap: 10px;
  margin-top: 18px;
}

.dialog-hint {
  margin: 0;
  color: #64748b;
  line-height: 1.5;
}

.doc-id {
  color: #1d4ed8;
}

.documents-scroll {
  max-height: 58vh;
  overflow-y: auto;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  background: #fff;
}

.documents-scroll :deep(.el-table__row) {
  cursor: pointer;
}

.json-preview {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 240px;
  overflow: auto;
  font-size: 12px;
}

.pagination-row {
  justify-content: flex-end;
  margin-top: 18px;
}

.selection-summary {
  margin-top: 12px;
  color: #64748b;
  font-size: 13px;
}

.complex-fields-tip {
  padding: 12px;
  background: #fff;
  border-radius: 12px;
}

.complex-fields-tip {
  color: #9a3412;
  margin-bottom: 12px;
}

.index-item {
  display: flex;
  align-items: center;
}

.index-list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 10px;
}

.index-item {
  justify-content: space-between;
  gap: 12px;
  padding: 12px;
  border-radius: 12px;
  background: #fff;
}

.index-item__main {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.index-item__main code {
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
  color: var(--primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.index-item__flags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.index-item__flags span {
  padding: 4px 8px;
  border-radius: 999px;
  background: #edf4ff;
  color: #365a7a;
  font-size: 12px;
}

.editor-meta {
  margin-bottom: 16px;
}

.editor-hint {
  color: #64748b;
  margin: 0;
}

.editor-toolbar {
  margin-bottom: 16px;
}

.form-editor {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-row {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr 1.2fr auto;
  gap: 10px;
}

.editor-surface {
  background: #0f172a;
  border-radius: 12px;
  padding: 12px;
}

.editor-surface :deep(.el-textarea__inner) {
  background: #0f172a;
  color: #e2e8f0;
  border: none;
  box-shadow: none;
  font-family: 'JetBrains Mono', 'Courier New', monospace;
}

.drawer-footer {
  justify-content: flex-end;
  width: 100%;
}

.hidden-input {
  display: none;
}

@media (max-width: 1200px) {
  .mongodb-console {
    grid-template-columns: 1fr;
  }

  .control-panels {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .hero-card,
  .workspace-header,
  .table-header,
  .table-header__right,
  .toolbar-actions,
  .hero-actions,
  .condition-row,
  .form-row,
  .topbar {
    display: flex;
    flex-direction: column;
    align-items: stretch;
  }

  .table-actions,
  .table-meta {
    justify-content: flex-start;
  }

  .hero-metrics,
  .stats-strip {
    grid-template-columns: 1fr;
  }

  .console-main {
    padding: 16px;
  }
}
</style>
