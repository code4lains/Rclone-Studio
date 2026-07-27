import { createPinia, defineStore } from 'pinia'

export interface Instance {
  id: string
  name: string
  type: string
  command: string[]
  url: string
  user: string
  pass: string
  pid: number
  status: string
}

export const useInstanceStore = defineStore('instances', {
  state: () => ({
    instances: [] as Instance[],
    isDesktop: !!window.go?.main?.App
  }),
  actions: {
    async fetchInstances() {
      if (this.isDesktop) {
        try {
          this.instances = await window.go!.main!.App!.GetInstances()
        } catch (e) {
          console.error(e)
        }
      } else {
        const stored = localStorage.getItem('rclone_instances')
        if (stored) {
          try {
            this.instances = JSON.parse(stored)
          } catch (e) {
            console.error(e)
          }
        }
      }
    },
    async saveInstance(inst: Instance) {
      if (this.isDesktop) {
        await window.go!.main!.App!.SaveInstance(inst)
      } else {
        const idx = this.instances.findIndex(i => i.id === inst.id)
        if (idx !== -1) {
          this.instances[idx] = inst
        } else {
          inst.status = 'unknown'
          this.instances.push(inst)
        }
        localStorage.setItem('rclone_instances', JSON.stringify(this.instances))
      }
      await this.fetchInstances()
    },
    async deleteInstance(id: string) {
      if (this.isDesktop) {
        await window.go!.main!.App!.DeleteInstance(id)
      } else {
        this.instances = this.instances.filter(i => i.id !== id)
        localStorage.setItem('rclone_instances', JSON.stringify(this.instances))
      }
      await this.fetchInstances()
    },
    async stopExternalInstanceWeb(inst: Instance) {
      if (this.isDesktop) return
      
      const u = new URL(inst.url)
      const port = u.port || (u.protocol === 'https:' ? '443' : '80')
      const url = `http://${u.hostname}:${port}/core/quit`
      
      const headers: HeadersInit = {
        'Content-Type': 'application/json',
      };

      if (inst.user || inst.pass) {
        const auth = btoa(`${inst.user}:${inst.pass}`);
        headers['Authorization'] = `Basic ${auth}`;
      }

      await fetch(url, {
        method: 'POST',
        headers,
        body: JSON.stringify({})
      })
    }
  }
})

const pinia = createPinia()

export const useConnectionStore = defineStore('connection', {
  state: () => ({
    isConnected: false,
    config: {
      ip: '',
      port: '',
      user: '',
      pass: ''
    }
  }),
  actions: {
    setConfig(config: { ip: string; port: string; user: string; pass: string }) {
      this.config = config
      localStorage.setItem('rclone_config', JSON.stringify(config))
    },
    setConnected(status: boolean) {
      this.isConnected = status
    },
    loadConfig() {
      const stored = localStorage.getItem('rclone_config')
      if (stored) {
        try {
          this.config = JSON.parse(stored)
        } catch (e) {}
      }
    }
  }
})

export default pinia
