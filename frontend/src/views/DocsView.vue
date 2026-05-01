<template>
  <div class="docs-view">
    <div class="docs-card">
      <p class="eyebrow">{{ copy.docs.eyebrow }}</p>
      <h1>{{ copy.docs.title }}</h1>
      <p class="docs-card__summary">{{ copy.docs.summary }}</p>
      <div class="docs-card__link-wrap">
        <el-link :href="openApiUrl" target="_blank" type="primary">{{ openApiUrl }}</el-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { copy } from '../copy'

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? window.location.origin
const openApiUrl = `${apiBaseUrl}/api/openapi.json`
</script>
