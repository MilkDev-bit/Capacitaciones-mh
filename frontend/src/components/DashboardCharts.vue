<template>
  <div class="dashboard-charts">
    <div class="chart-card">
      <h4>Actividad Reciente</h4>
      <div class="chart-wrapper">
        <Bar :data="barData" :options="barOptions" />
      </div>
    </div>
    <div class="chart-card">
      <h4>Estado de Aprobación</h4>
      <div class="chart-wrapper">
        <Doughnut :data="doughnutData" :options="doughnutOptions" />
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
import { onMounted, onUnmounted, ref, computed } from 'vue'

const textColor = ref('#86868b')
const gridColor = ref('rgba(0,0,0,0.05)')
const surfaceColor = ref('#ffffff')

const updateChartColors = () => {
  const isDark = document.documentElement.classList.contains('dark-theme') || 
                (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches && !document.documentElement.classList.contains('light-theme'))
  textColor.value = isDark ? '#aeaeb2' : '#86868b'
  gridColor.value = isDark ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.05)'
  surfaceColor.value = isDark ? '#1d1d1f' : '#ffffff'
}

const observer = new MutationObserver(updateChartColors)

onMounted(() => {
  updateChartColors()
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] })
})
onUnmounted(() => {
  observer.disconnect()
})

const barOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: surfaceColor.value === '#ffffff' ? 'rgba(0,0,0,0.8)' : 'rgba(255,255,255,0.9)',
      titleColor: surfaceColor.value === '#ffffff' ? '#ffffff' : '#000000',
      bodyColor: surfaceColor.value === '#ffffff' ? '#ffffff' : '#000000',
      padding: 12,
      cornerRadius: 8,
      displayColors: false,
    }
  },
  scales: {
    x: {
      ticks: { color: textColor.value, font: { family: 'Inter', size: 12 } },
      grid: { display: false },
      border: { display: false }
    },
    y: {
      ticks: { color: textColor.value, font: { family: 'Inter', size: 12 } },
      grid: { color: gridColor.value, borderDash: [5, 5] },
      border: { display: false }
    }
  }
}))

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '75%',
  plugins: {
    legend: {
      position: 'bottom' as const,
      labels: {
        color: textColor.value,
        padding: 20,
        font: { family: 'Inter', size: 13, weight: 'bold' as const },
        usePointStyle: true,
        pointStyle: 'circle'
      }
    },
    tooltip: {
      backgroundColor: surfaceColor.value === '#ffffff' ? 'rgba(0,0,0,0.8)' : 'rgba(255,255,255,0.9)',
      titleColor: surfaceColor.value === '#ffffff' ? '#ffffff' : '#000000',
      bodyColor: surfaceColor.value === '#ffffff' ? '#ffffff' : '#000000',
      padding: 12,
      cornerRadius: 8
    }
  }
}))

const barData = computed(() => ({
  labels: ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun'],
  datasets: [
    {
      label: 'Accesos',
      backgroundColor: '#f97316',
      hoverBackgroundColor: '#ea580c',
      borderRadius: 6,
      barThickness: 32,
      data: [40, 55, 45, 70, 90, 85]
    }
  ]
}))

const doughnutData = computed(() => ({
  labels: ['Aprobados', 'Reprobados', 'En Progreso'],
  datasets: [
    {
      backgroundColor: ['#34c759', '#ff3b30', '#f59e0b'],
      hoverBackgroundColor: ['#28a745', '#dc3545', '#e0a800'],
      borderWidth: 3,
      borderColor: surfaceColor.value,
      hoverBorderColor: surfaceColor.value,
      data: [65, 15, 20]
    }
  ]
}))
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
