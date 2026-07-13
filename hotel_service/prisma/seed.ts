import 'dotenv/config';
import { Pool } from 'pg';
import { PrismaPg } from '@prisma/adapter-pg';
import { PrismaClient } from '../src/infra/database/generated/client.js';

const connectionString = `${process.env.DATABASE_URL}`;

const pool = new Pool({ connectionString });
const adapter = new PrismaPg(pool);
const prisma = new PrismaClient({ adapter });

// Schema has one `State` model with no type discriminator — Union
// Territories are seeded as rows in the same table, exactly as the
// schema models them today.
type StateSeed = {
    name: string;
    cities: string[];
};

const STATES_AND_UTS: StateSeed[] = [
    // ---- States (28) ----
    { name: 'Andhra Pradesh', cities: ['Visakhapatnam', 'Vijayawada'] },
    { name: 'Arunachal Pradesh', cities: ['Itanagar'] },
    { name: 'Assam', cities: ['Guwahati', 'Dibrugarh'] },
    { name: 'Bihar', cities: ['Patna', 'Gaya'] },
    { name: 'Chhattisgarh', cities: ['Raipur'] },
    { name: 'Goa', cities: ['Panaji'] },
    { name: 'Gujarat', cities: ['Ahmedabad', 'Surat'] },
    { name: 'Haryana', cities: ['Gurugram', 'Faridabad'] },
    { name: 'Himachal Pradesh', cities: ['Shimla'] },
    { name: 'Jharkhand', cities: ['Ranchi'] },
    { name: 'Karnataka', cities: ['Bengaluru', 'Mysuru'] },
    { name: 'Kerala', cities: ['Kochi', 'Thiruvananthapuram'] },
    { name: 'Madhya Pradesh', cities: ['Bhopal', 'Indore'] },
    { name: 'Maharashtra', cities: ['Mumbai', 'Pune'] },
    { name: 'Manipur', cities: ['Imphal'] },
    { name: 'Meghalaya', cities: ['Shillong'] },
    { name: 'Mizoram', cities: ['Aizawl'] },
    { name: 'Nagaland', cities: ['Kohima'] },
    { name: 'Odisha', cities: ['Bhubaneswar'] },
    { name: 'Punjab', cities: ['Amritsar', 'Ludhiana'] },
    { name: 'Rajasthan', cities: ['Jaipur', 'Udaipur'] },
    { name: 'Sikkim', cities: ['Gangtok'] },
    { name: 'Tamil Nadu', cities: ['Chennai', 'Coimbatore'] },
    { name: 'Telangana', cities: ['Hyderabad'] },
    { name: 'Tripura', cities: ['Agartala'] },
    { name: 'Uttar Pradesh', cities: ['Lucknow', 'Agra'] },
    { name: 'Uttarakhand', cities: ['Dehradun'] },
    { name: 'West Bengal', cities: ['Kolkata', 'Darjeeling'] },

    // ---- Union Territories (8) ----
    { name: 'Andaman and Nicobar Islands', cities: ['Port Blair'] },
    { name: 'Chandigarh', cities: ['Chandigarh'] },
    {
        name: 'Dadra and Nagar Haveli and Daman and Diu',
        cities: ['Daman'],
    },
    { name: 'Delhi', cities: ['New Delhi'] },
    { name: 'Jammu and Kashmir', cities: ['Srinagar', 'Jammu'] },
    { name: 'Ladakh', cities: ['Leh'] },
    { name: 'Lakshadweep', cities: ['Kavaratti'] },
    { name: 'Puducherry', cities: ['Puducherry'] },
];

const ROOM_CATEGORIES = [
    {
        roomType: 'STANDARD' as const,
        description:
            'Basic amenities, smallest room size — our most budget-friendly option.',
    },
    {
        roomType: 'DELUXE' as const,
        description:
            'Mid-tier — more space and a better view/amenities than Standard.',
    },
    {
        roomType: 'SUITE' as const,
        description:
            'Our most premium option — multiple rooms/living area with top-tier amenities.',
    },
];

async function seedStatesAndCities(): Promise<void> {
    for (const state of STATES_AND_UTS) {
        await prisma.state.create({
            data: {
                name: state.name,
                cities: {
                    create: state.cities.map((cityName) => ({
                        name: cityName,
                    })),
                },
            },
        });
    }
    console.log(
        `Seeded ${STATES_AND_UTS.length} states/UTs and ${STATES_AND_UTS.reduce(
            (sum, s) => sum + s.cities.length,
            0
        )} cities`
    );
}

async function seedRoomCategories(): Promise<void> {
    await prisma.roomCategory.createMany({
        data: ROOM_CATEGORIES,
    });
    console.log(`Seeded ${ROOM_CATEGORIES.length} room categories`);
}

async function main(): Promise<void> {
    await seedStatesAndCities();
    await seedRoomCategories();
}

main()
    .catch((err) => {
        console.error('Seed failed:', err);
        process.exit(1);
    })
    .finally(async () => {
        await prisma.$disconnect();
        await pool.end();
    });
