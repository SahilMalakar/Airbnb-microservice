export function calculateNights(checkInDate: Date, checkOutDate: Date): number {
    return Math.round(
        (checkOutDate.getTime() - checkInDate.getTime()) / (1000 * 60 * 60 * 24)
    );
}