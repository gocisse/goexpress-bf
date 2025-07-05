#!/bin/bash

# Optimized build script for low memory servers
echo "🚀 Starting optimized build for low memory server..."
echo "=================================================="

# Check available memory
echo "💾 Available memory:"
free -h

# Set Node.js memory limits
export NODE_OPTIONS="--max-old-space-size=512"

# Clear any existing build
echo "🧹 Cleaning previous build..."
rm -rf dist .vite node_modules/.vite

# Build with memory optimization
echo "🔨 Building with memory optimization..."
echo "This may take 3-5 minutes on a 1GB server..."

# Run TypeScript compilation first (less memory intensive)
echo "📝 Running TypeScript compilation..."
npx tsc --noEmit

if [ $? -eq 0 ]; then
    echo "✅ TypeScript compilation successful"
    
    # Run Vite build with memory limits
    echo "📦 Running Vite build..."
    npx vite build --mode production
    
    if [ $? -eq 0 ]; then
        echo "✅ Build completed successfully!"
        echo "📊 Build size:"
        du -sh dist/
        echo ""
        echo "🎉 Ready to start with: npm run preview"
    else
        echo "❌ Vite build failed"
        exit 1
    fi
else
    echo "❌ TypeScript compilation failed"
    exit 1
fi

