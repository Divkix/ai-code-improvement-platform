// MongoDB initialization script to create demo user
// This script runs automatically when MongoDB container starts for the first time

print('🚀 Initializing MongoDB for ACIP...');

// Switch to the acip database
db = db.getSiblingDB('acip');

// Create demo user with hashed password (bcrypt hash of "demo123456")
// Hash generated using Go bcrypt.GenerateFromPassword with DefaultCost (10)
const demoUser = {
    email: 'demo@acip.com',
    password: '$2a$10$Q2kiuLcrSrXLHLBkEoJIz.ZeBfQE5iEt6vqgNQDC0PiIgeGVO0I9O',
    name: 'Demo User',
    createdAt: new Date(),
    updatedAt: new Date()
};

// Check if demo user already exists
const existingUser = db.users.findOne({ email: demoUser.email });

if (existingUser) {
    print('🌱 Demo user already exists, skipping creation');
} else {
    // Insert demo user
    const result = db.users.insertOne(demoUser);
    if (result.acknowledged) {
        print('🌱 Created demo user: ' + demoUser.name + ' (' + demoUser.email + ')');
    } else {
        print('❌ Failed to create demo user');
        throw new Error('Failed to insert demo user into the database');
    }
    print('📧 Login with: demo@acip.com / demo123456');
}

// Create index on email for better performance
db.users.createIndex({ email: 1 }, { unique: true });
print('📝 Created unique index on users.email');

print('✅ MongoDB initialization completed successfully');