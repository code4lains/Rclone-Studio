// Declare window.go for TypeScript
declare global {
  interface Window {
    go?: {
      main?: {
        App?: {
          CallRcloneAPI(endpoint: string, payload: any, ip: string, port: string, user: string, pass: string): Promise<any>;
          GetInstances(): Promise<any[]>;
          SaveInstance(inst: any): Promise<void>;
          DeleteInstance(id: string): Promise<void>;
          StartInstance(id: string): Promise<void>;
          StopInstance(id: string): Promise<void>;
          RestartInstance(id: string): Promise<void>;
        }
      }
    }
  }
}

/**
 * Call Rclone API with a unified bridge for both Wails (Desktop) and Web environments.
 * @param endpoint The Rclone rc API endpoint (e.g., 'core/version')
 * @param payload The JSON payload to send
 * @returns The JSON response from Rclone
 */
export async function callRcloneAPI(endpoint: string, payload: any = {}, silent: boolean = false): Promise<any> {
  const rcloneConfigStr = localStorage.getItem('rclone_config');
  let ip = '';
  let port = '';
  let user = '';
  let pass = '';

  if (rcloneConfigStr) {
    try {
      const config = JSON.parse(rcloneConfigStr);
      ip = config.ip || ip;
      port = config.port || port;
      user = config.user || user;
      pass = config.pass || pass;
    } catch (e) {
      if (!silent) console.warn('Failed to parse rclone_config from localStorage');
    }
  }

  // Check if we are in the Wails environment
  if (window.go?.main?.App?.CallRcloneAPI) {
    try {
      const response = await window.go.main.App.CallRcloneAPI(endpoint, payload, ip, port, user, pass);
      // Wails backend will return an object or throw an error string
      return response;
    } catch (error) {
      if (!silent) console.error('Wails API Error:', error);
      throw error;
    }
  }


  const url = `http://${ip}:${port}/${endpoint}`;
  
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (user || pass) {
    const auth = btoa(`${user}:${pass}`);
    headers['Authorization'] = `Basic ${auth}`;
  }

  try {
    const response = await fetch(url, {
      method: 'POST',
      headers,
      body: JSON.stringify(payload)
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return await response.json();
  } catch (error) {
    if (!silent) console.error('Web Fetch API Error:', error);
    throw error;
  }
}
