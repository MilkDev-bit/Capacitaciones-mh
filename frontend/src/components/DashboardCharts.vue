<template>
  <div class="dashboard-charts">
    <div class="chart-card">
      <h4>Asignaciones (Año Actual)</h4>
      <div class="chart-wrapper">
        <Bar :data="barData" :options="barOptions" />
      </div>
    </div>
    <div class="chart-card">
      <h4>Tipos de Usuario</h4>
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
  CategoryScale,
  LinearScale,
  ArcElement,
  BarElement
} from 'chart.js'
import { Bar, Doughnut } from 'vue-chartjs'
import { computed } from 'vue'

const props = defineProps<{
  users: any[]
  asignaciones: any[]
}>()
import { useTheme } from '@/composables/useTheme'
ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend, ArcElement)
const { isDark } = useTheme()

const textColor = computed(() => isDark.value ? '#aeaeb2' : '#86868b')
const gridColor = computed(() => isDark.value ? 'rgba(255,255,255,0.05)' : 'rgba(0,0,0,0.05)')
const surfaceColor = computed(() => isDark.value ? '#1d1d1f' : '#ffffff')

const barOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: {
      backgroundColor: isDark.value ? 'rgba(255,255,255,0.9)' : 'rgba(0,0,0,0.8)',
      titleColor: isDark.value ? '#000000' : '#ffffff',
      bodyColor: isDark.value ? '#000000' : '#ffffff',
      padding: 12,
      cornerRadius: 8,
      displayColors: true,
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
      backgroundColor: isDark.value ? 'rgba(255,255,255,0.9)' : 'rgba(0,0,0,0.8)',
      titleColor: isDark.value ? '#000000' : '#ffffff',
      bodyColor: isDark.value ? '#000000' : '#ffffff',
      padding: 12,
      cornerRadius: 8
    }
  }
}))

const barData = computed(() => {
  const months = ['Ene', 'Feb', 'Mar', 'Abr', 'May', 'Jun', 'Jul', 'Ago', 'Sep', 'Oct', 'Nov', 'Dic']
  const counts = new Array(12).fill(0)
  
  const currentYear = new Date().getFullYear()
  const currentMonth = new Date().getMonth()
  
  if (props.asignaciones) {
    props.asignaciones.forEach(a => {
      const d = new Date(a.assigned_at)
      if (d.getFullYear() === currentYear) {
        counts[d.getMonth()]++
      }
    })
  }

  const dataLength = Math.max(currentMonth + 1, 1)

  return {
    labels: months.slice(0, dataLength),
    datasets: [
      {
        label: 'Asignaciones Nuevas',
        backgroundColor: (context: any) => {
          const chart = context.chart;
          if (!chart) return '#f97316';
          const ctx = chart.ctx;
          if (!ctx) return '#f97316';
          const gradient = ctx.createLinearGradient(0, 0, 0, 300);
          gradient.addColorStop(0, '#f97316');
          gradient.addColorStop(1, isDark.value ? 'rgba(249,115,22,0.15)' : 'rgba(249,115,22,0.05)');
          return gradient;
        },
        hoverBackgroundColor: '#ea580c',
        borderColor: '#f97316',
        borderWidth: 2,
        borderRadius: 6,
        barThickness: 24,
        data: counts.slice(0, dataLength)
      }
    ]
  }
})

const doughnutData = computed(() => {
  let admins = 0
  let insts = 0
  let studs = 0

  if (props.users) {
    admins = props.users.filter(u => u.role === 'admin').length
    insts = props.users.filter(u => u.role === 'instructor').length
    studs = props.users.filter(u => u.role === 'user').length
  }

  return {
    labels: ['Estudiantes', 'Instructores', 'Admins'],
    datasets: [
      {
        backgroundColor: ['#3b82f6', '#f59e0b', '#8b5cf6'],
        hoverBackgroundColor: ['#2563eb', '#d97706', '#7c3aed'],
        borderWidth: 4,
        borderColor: surfaceColor.value,
        hoverBorderColor: surfaceColor.value,
        data: [studs, insts, admins]
      }
    ]
  }
})
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
