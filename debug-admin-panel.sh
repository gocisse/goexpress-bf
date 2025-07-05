#!/bin/bash

# Debug GoExpress Admin Panel Access Issues
echo "🔍 Debugging GoExpress Admin Panel Access"
echo "========================================="

SERVER_IP="144.21.63.195"
ADMIN_PORT="3001"

echo "🌐 Server IP: $SERVER_IP"
echo "📊 Admin Port: $ADMIN_PORT"
echo ""

# Test 1: Check if service is running
echo "1. Service Status Check"
echo "----------------------"
sudo systemctl status goexpress-admin --no-pager -l
echo ""

# Test 2: Check if port is listening
echo "2. Port Listening Check"
echo "----------------------"
echo "Checking if port $ADMIN_PORT is listening:"
sudo netstat -tlnp | grep ":$ADMIN_PORT "
echo ""

# Test 3: Check process details
echo "3. Process Details"
echo "-----------------"
echo "GoExpress admin processes:"
ps aux | grep -E "(vite|npm|node)" | grep -v grep
echo ""

# Test 4: Check logs
echo "4. Service Logs (last 20 lines)"
echo "-------------------------------"
sudo journalctl -u goexpress-admin -n 20 --no-pager
echo ""

# Test 5: Test local access
echo "5. Local Access Test"
echo "-------------------"
echo "Testing local access to admin panel:"
if curl -s --connect-timeout 5 http://localhost:$ADMIN_PORT > /dev/null; then
    echo "✅ Local access works"
else
    echo "❌ Local access failed"
fi
echo ""

# Test 6: Check if vite is binding to correct interface
echo "6. Network Interface Binding"
echo "---------------------------"
echo "Checking what interfaces the service is bound to:"
sudo lsof -i :$ADMIN_PORT
echo ""

# Test 7: Check firewall status
echo "7. Firewall Status"
echo "-----------------"
echo "UFW Status:"
sudo ufw status
echo ""
echo "iptables rules:"
sudo iptables -L INPUT -n | grep $ADMIN_PORT || echo "No iptables rules found for port $ADMIN_PORT"
echo ""

# Test 8: Manual curl test with verbose output
echo "8. Detailed Connection Test"
echo "--------------------------"
echo "Testing connection with verbose output:"
curl -v --connect-timeout 10 http://localhost:$ADMIN_PORT 2>&1 | head -20
echo ""

# Test 9: Check if build exists
echo "9. Build Directory Check"
echo "-----------------------"
if [ -d "dist" ]; then
    echo "✅ Build directory exists"
    echo "Build contents:"
    ls -la dist/ | head -10
else
    echo "❌ Build directory missing"
fi
echo ""

# Test 10: Check package.json scripts
echo "10. Package.json Scripts"
echo "-----------------------"
echo "Available scripts:"
grep -A 10 '"scripts"' package.json
echo ""

echo "🔧 Troubleshooting Summary"
echo "========================="
echo "If the service is running but not accessible:"
echo "1. Check if vite is binding to 0.0.0.0 (all interfaces)"
echo "2. Verify the build was successful"
echo "3. Check for any error messages in logs"
echo "4. Ensure Oracle Cloud Security List allows port $ADMIN_PORT"
echo ""
echo "🚀 Quick fixes to try:"
echo "sudo systemctl restart goexpress-admin"
echo "npm run build && npm run preview"
echo "sudo journalctl -u goexpress-admin -f"

