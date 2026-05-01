<template>
  <div class="connect-shell">
    <section class="connect-card">
      <div class="connect-card__intro">
        <p class="eyebrow">MongoDB</p>
        <h1>{{ t.connectTitle }}</h1>
        <p class="connect-summary">{{ t.connectSummary }}</p>
      </div>

      <el-form label-position="top" class="connect-form">
        <div class="connect-grid">
          <el-form-item :label="t.host">
            <el-input v-model="form.host" />
            <div class="field-hint">{{ t.hostHint }}</div>
          </el-form-item>
          <el-form-item :label="t.port">
            <el-input-number v-model="form.port" :min="1" :max="65535" class="connect-number" />
          </el-form-item>
        </div>

        <div class="connect-grid">
          <el-form-item :label="t.database">
            <el-input v-model="form.database" />
            <div class="field-hint">{{ t.databaseOptionalHint }}</div>
          </el-form-item>
          <el-form-item :label="t.authSource">
            <el-input v-model="form.authSource" />
            <div class="field-hint">{{ t.authSourceHint }}</div>
          </el-form-item>
        </div>

        <div class="connect-grid">
          <el-form-item :label="t.username">
            <el-input v-model="form.username" />
          </el-form-item>
          <el-form-item :label="t.password">
            <el-input v-model="form.password" type="password" show-password @keyup.enter="connectAndEnter" />
          </el-form-item>
        </div>

        <div class="session-hint">{{ t.sessionScopeHint }}</div>

        <div class="connect-actions">
          <el-button :loading="testing" @click="runConnectionTest">{{ t.testConnection }}</el-button>
          <el-button type="primary" :loading="connecting" @click="connectAndEnter">{{ t.connectAndEnter }}</el-button>
        </div>
      </el-form>
    </section>

    <aside class="connect-sidecard">
      <div class="connect-sidecard__header">
        <p class="eyebrow">{{ t.currentTarget }}</p>
        <strong>{{ sessionConnectionLabel }}</strong>
      </div>
      <div class="target-list">
        <div class="target-item">
          <span>{{ t.host }}</span>
          <strong>{{ sessionConnection?.host || '-' }}</strong>
        </div>
        <div class="target-item">
          <span>{{ t.port }}</span>
          <strong>{{ sessionConnection?.port || '-' }}</strong>
        </div>
        <div class="target-item">
          <span>{{ t.currentTargetDatabase }}</span>
          <strong>{{ sessionConnection?.database || '-' }}</strong>
        </div>
        <div class="target-item">
          <span>{{ t.username }}</span>
          <strong>{{ sessionConnection?.username || '-' }}</strong>
        </div>
      </div>
      <div class="connect-sidecard__actions">
        <el-button plain :disabled="!sessionConnection" @click="enterWorkspace">{{ t.switchConnection }}</el-button>
        <el-button plain type="danger" :disabled="!sessionConnection" @click="disconnect">{{ t.disconnect }}</el-button>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

import { fetchConnectionInfo, testConnection } from '../api/mongo'
import { useMongoI18n } from '../i18n/mongodb'
import {
  clearMongoSessionConnection,
  defaultMongoConnection,
  getMongoSessionConnection,
  setMongoSessionConnection,
  type MongoSessionConnection,
} from '../session/mongoConnection'

const router = useRouter()
const route = useRoute()
const { t } = useMongoI18n()

const sessionConnection = ref<MongoSessionConnection | null>(getMongoSessionConnection())
const testing = ref(false)
const connecting = ref(false)

const form = reactive<MongoSessionConnection>(sessionConnection.value || defaultMongoConnection())

const sessionConnectionLabel = computed(() => {
  if (!sessionConnection.value) {
    return t.value.sessionConnectionRequired
  }

  return `${sessionConnection.value.host}:${sessionConnection.value.port}`
})

async function runConnectionTest() {
  testing.value = true
  try {
    await testConnection({ ...form })
    ElMessage.success(t.value.connectionTestSuccess)
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    testing.value = false
  }
}

async function connectAndEnter() {
  connecting.value = true
  try {
    await testConnection({ ...form })
    setMongoSessionConnection({ ...form })
    sessionConnection.value = getMongoSessionConnection()
    await fetchConnectionInfo()
    ElMessage.success(t.value.connectionSaved)
    await enterWorkspace()
  } catch (error) {
    ElMessage.error((error as Error).message)
  } finally {
    connecting.value = false
  }
}

async function enterWorkspace() {
  const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
  await router.push(redirect)
}

function disconnect() {
  clearMongoSessionConnection()
  sessionConnection.value = null
  Object.assign(form, defaultMongoConnection())
}
</script>

<style scoped>
.connect-shell {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(0, 720px) 320px;
  gap: 24px;
  padding: 32px;
  background:
    radial-gradient(circle at top left, rgba(57, 124, 255, 0.16), transparent 32%),
    linear-gradient(180deg, #f6f9ff 0%, #eef3f8 100%);
}

.connect-card,
.connect-sidecard {
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid rgba(219, 228, 240, 0.9);
  border-radius: 24px;
  box-shadow: 0 24px 64px rgba(15, 23, 42, 0.08);
}

.connect-card {
  padding: 32px;
}

.connect-card__intro {
  margin-bottom: 24px;
}

.eyebrow {
  margin: 0 0 8px;
  color: #7c8aa5;
  font-size: 12px;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.connect-card h1,
.connect-sidecard strong {
  font-family: "Manrope", sans-serif;
}

.connect-card h1 {
  margin: 0 0 12px;
  font-size: clamp(2rem, 3vw, 2.8rem);
  color: #162033;
}

.connect-summary,
.field-hint,
.session-hint {
  color: #60708c;
  line-height: 1.6;
}

.connect-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.connect-number {
  width: 100%;
}

.field-hint {
  margin-top: 8px;
  font-size: 12px;
}

.session-hint {
  margin-top: 4px;
  font-size: 13px;
}

.connect-actions {
  display: flex;
  gap: 12px;
  margin-top: 24px;
}

.connect-sidecard {
  padding: 24px;
  align-self: start;
}

.connect-sidecard__header {
  margin-bottom: 16px;
}

.connect-sidecard__header strong {
  display: block;
  font-size: 1.3rem;
  color: #162033;
}

.target-list {
  display: grid;
  gap: 12px;
}

.target-item {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 14px;
  border-radius: 14px;
  background: #f7f9fc;
  color: #334155;
}

.target-item span {
  color: #64748b;
}

.connect-sidecard__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 20px;
}

@media (max-width: 960px) {
  .connect-shell {
    grid-template-columns: 1fr;
    padding: 20px;
  }

  .connect-grid {
    grid-template-columns: 1fr;
  }
}
</style>
