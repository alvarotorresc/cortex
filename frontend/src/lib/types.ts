export interface PluginManifest {
  id: string;
  name: string;
  version: string;
  description: string;
  icon: string;
  color: string;
  permissions: string[];
}

export interface WidgetLayout {
  id: number;
  widget_id: string;
  position_x: number;
  position_y: number;
  width: number;
  height: number;
  created_at: string;
  updated_at: string;
}
