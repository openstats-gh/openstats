import { PUBLIC_OPENSTATS_API_URL } from "$env/static/public";

interface OpenstatsConfig {
    apiBaseUrl: string
}

export default {
    apiBaseUrl: PUBLIC_OPENSTATS_API_URL
} as OpenstatsConfig;