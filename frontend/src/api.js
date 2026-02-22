// API service for communicating with the backend

const API_BASE_URL = '/api';

// Helper function for making requests
async function request(url, options = {}) {
  const response = await fetch(`${API_BASE_URL}${url}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(`API Error: ${response.status} - ${errorText}`);
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return null;
  }

  const data = await response.json();
  // Handle null responses from Go (nil slices encode as null)
  return data;
}

// Device API
export const deviceAPI = {
  async list() {
    // Note: Backend may need GET /api/devices endpoint for listing all devices
    return request('/devices');
  },

  async get(deviceId) {
    return request(`/devices/${deviceId}`);
  },

  async create(device, subsystemId) {
    return request('/devices', {
      method: 'POST',
      body: JSON.stringify({ device, subsystem_id: subsystemId }),
    });
  },

  async update(deviceId, device) {
    return request(`/devices/${deviceId}`, {
      method: 'PUT',
      body: JSON.stringify(device),
    });
  },

  async delete(deviceId) {
    return request(`/devices/${deviceId}`, {
      method: 'DELETE',
    });
  },
};

// Sensor API
export const sensorAPI = {
  async list(params = {}) {
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        queryParams.append(key, value);
      }
    });
    const query = queryParams.toString();
    return request(`/sensors${query ? '?' + query : ''}`);
  },

  async get(deviceId, sensorId) {
    return request(`/sensors/${deviceId}/${sensorId}`);
  },

  async create(sensor) {
    return request('/sensors', {
      method: 'POST',
      body: JSON.stringify(sensor),
    });
  },

  async update(deviceId, sensorId, sensor) {
    return request(`/sensors/${deviceId}/${sensorId}`, {
      method: 'PUT',
      body: JSON.stringify(sensor),
    });
  },

  async delete(deviceId, sensorId) {
    return request(`/sensors/${deviceId}/${sensorId}`, {
      method: 'DELETE',
    });
  },

  // Sensor Reading API
  async getReadings(params = {}) {
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        queryParams.append(key, value);
      }
    });
    const query = queryParams.toString();
    return request(`/sensor-readings${query ? '?' + query : ''}`);
  },

  async getLatest(sensorId) {
    try {
      return await request(`/sensor-readings/${sensorId}/latest`);
    } catch (err) {
      // Return null if no readings exist (404)
      if (err.message?.includes('404')) {
        return null;
      }
      throw err;
    }
  },

  async createReading(reading) {
    return request('/sensor-readings', {
      method: 'POST',
      body: JSON.stringify(reading),
    });
  },
};

// Actuator API
export const actuatorAPI = {
  async list(params = {}) {
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        queryParams.append(key, value);
      }
    });
    const query = queryParams.toString();
    return request(`/actuators${query ? '?' + query : ''}`);
  },

  async get(deviceId, actuatorId) {
    return request(`/actuators/${deviceId}/${actuatorId}`);
  },

  async create(actuator) {
    return request('/actuators', {
      method: 'POST',
      body: JSON.stringify(actuator),
    });
  },

  async update(deviceId, actuatorId, actuator) {
    return request(`/actuators/${deviceId}/${actuatorId}`, {
      method: 'PUT',
      body: JSON.stringify(actuator),
    });
  },

  async delete(deviceId, actuatorId) {
    return request(`/actuators/${deviceId}/${actuatorId}`, {
      method: 'DELETE',
    });
  },

  // Actuator State API
  async getStates(params = {}) {
    const queryParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        queryParams.append(key, value);
      }
    });
    const query = queryParams.toString();
    return request(`/actuator-states${query ? '?' + query : ''}`);
  },

  async getLatest(actuatorId) {
    try {
      return await request(`/actuator-states/${actuatorId}/latest`);
    } catch (err) {
      // Return null if no states exist (404)
      if (err.message?.includes('404')) {
        return null;
      }
      throw err;
    }
  },

  async getStatusByTag(tag) {
    try {
      return await request(`/actuators/by-tag/${encodeURIComponent(tag)}/status`);
    } catch (err) {
      // Return null if no status exists (404)
      if (err.message?.includes('404')) {
        return null;
      }
      throw err;
    }
  },

  async createState(state) {
    return request('/actuator-states', {
      method: 'POST',
      body: JSON.stringify(state),
    });
  },
};

// Workflow API
export const workflowAPI = {
  async startDiscovery(options = {}) {
    return request('/workflows/discovery', {
      method: 'POST',
      body: JSON.stringify({ options }),
    });
  },

  async getStatus(workflowId) {
    return request(`/workflows/${workflowId}`);
  },

  async list() {
    return request('/workflows');
  },
};
