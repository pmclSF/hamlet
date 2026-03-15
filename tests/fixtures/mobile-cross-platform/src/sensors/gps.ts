export interface Location {
  latitude: number;
  longitude: number;
  accuracy: number;
}

export function isValidLocation(loc: Location): boolean {
  return loc.latitude >= -90 && loc.latitude <= 90
    && loc.longitude >= -180 && loc.longitude <= 180
    && loc.accuracy > 0;
}

export function distanceBetween(a: Location, b: Location): number {
  const dlat = b.latitude - a.latitude;
  const dlon = b.longitude - a.longitude;
  return Math.sqrt(dlat * dlat + dlon * dlon) * 111000;
}
