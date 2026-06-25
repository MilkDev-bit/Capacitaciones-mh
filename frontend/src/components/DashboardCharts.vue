<template>
  <div class="dashboard-charts">
    <div class="chart-card">
      <h4>Actividad Reciente</h4>
      <div class="chart-wrapper">
        <Bar :data="barData" :options="chartOptions" />
      </div>
    </div>
    <div class="chart-card">
      <h4>Estado de Aprobación</h4>
      <div class="chart-wrapper">
        <Doughnut :data="doughnutData" :options="chartOptions" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  BarElement,
  CategoryScale,
  LinearScale,
  ArcElement
} from 'chart.js'
import { Bar, Doughnut } from 'vue-chartjs'

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend, ArcElement)

// To support dark mode nicely
import { onMounted, onUnmounted, ref } from 'vue'

const textColor = ref('#86868b')
const gridColor = ref('rgba(0,0,0,0.05)')

const updateChartColors = () => {
  const isDark = document.documentElement.classList.contains('dark-theme') || 
                (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches && !document.documentElement.classList.contains('light-theme'))
  textColor.value = isDark ? '#aeaeb2' : '#86868b'
  gridColor.value = isDark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.05)'
}

const observer = new MutationObserver(updateChartColors)

onMounted(() => {
  updateChartColors()
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})
onUnmounted(() => {
  observer.disconnect()
})

const chartOptions = ref({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        color: textColor
      }
    }
  },
  scales: {
    x: {
      ticks: { color: textColor },
      grid: { display: false }
    },
    y: {
      ticks: { color: textColor },
      grid: { color: gridColor }
    }
  }
})

const barData = {
  labels: ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun'],
  datasets: [
    {
      label: 'Accesos',
      backgroundColor: '#f97316',
      borderRadius: 4,
      data: [40, 55, 45, 70, 90, 85]
    }
  ]
}

const doughnutData = {
  labels: ['Aprobados', 'Reprobados', 'En Progreso'],
  datasets: [
    {
      backgroundColor: ['#22c55e', '#ef4444', '#f59e0b'],
      borderWidth: 0,
      data: [65, 15, 20]
    }
  ]
}
</script>

<style scoped>
.dashboard-charts {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 24px;
  margin-bottom: 24px;
}
.chart-card {
  background: var(--surface);
  border-radius: var(--r-lg);
  padding: 24px;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-light);
  height: 340px;
  display: flex;
  flex-direction: column;
}
.chart-card h4 {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--dark);
  margin-bottom: 16px;
}
.chart-wrapper {
  flex: 1;
  position: relative;
  min-height: 0;
}
</style>
