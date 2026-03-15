export interface CameraConfig {
  resolution: '720p' | '1080p' | '4k';
  facing: 'front' | 'rear';
  flash: boolean;
}

export function createCameraConfig(facing: 'front' | 'rear'): CameraConfig {
  return { resolution: '1080p', facing, flash: facing === 'rear' };
}

export function validateResolution(resolution: string): boolean {
  return ['720p', '1080p', '4k'].includes(resolution);
}
