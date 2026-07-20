<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getClusterSummary, type ClusterSummary } from './api/cluster'

const loading = ref(true)
const error = ref('')
const summary = ref<ClusterSummary | null>(null)

onMounted(async () => {
  try {
    summary.value = await getClusterSummary()
  } catch (err) {
    error.value = err instanceof Error ? err.message : '请求集群摘要失败'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <main class="shell">
    <aside class="sidebar">
      <div class="brand">
        <img class="brand-mark" src="/ceph-tower-logo.svg" alt="CephTower logo" />
        <div>
          <strong>CephTower</strong>
          <small>Ceph 管理控制台</small>
        </div>
      </div>
      <nav>
        <a class="active" href="#">集群概览</a>
        <a href="#">存储池</a>
        <a href="#">OSD</a>
        <a href="#">监控</a>
      </nav>
    </aside>

    <section class="content">
      <header class="topbar">
        <div>
          <h1>集群概览</h1>
          <p>通过 Ceph Manager Dashboard API 汇总集群运行状态。</p>
        </div>
      </header>

      <div v-if="loading" class="state">正在加载...</div>
      <div v-else-if="error" class="state error">{{ error }}</div>
      <div v-else class="metrics">
        <article class="metric">
          <span>健康状态</span>
          <strong>{{ summary?.health_status ?? 'unknown' }}</strong>
        </article>
        <article class="metric">
          <span>Ceph 版本</span>
          <strong>{{ summary?.version || '未配置' }}</strong>
        </article>
      </div>
    </section>
  </main>
</template>
