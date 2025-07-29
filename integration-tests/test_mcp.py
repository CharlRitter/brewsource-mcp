#!/usr/bin/env python3
"""
Simple MCP client to test Phase 1 functionality of BrewSource MCP Server.
Tests the three core tools: bjcp_lookup, search_beers, find_breweries
"""

import asyncio
import websockets
import json
import sys

async def test_mcp_server():
    uri = "ws://localhost:8080/mcp"

    try:
        async with websockets.connect(uri) as websocket:
            print("🍺 Connected to BrewSource MCP Server")
            print("=" * 50)

            # Test 1: Initialize MCP connection
            init_request = {
                "jsonrpc": "2.0",
                "id": 1,
                "method": "initialize",
                "params": {
                    "protocolVersion": "2024-11-05",
                    "capabilities": {
                        "roots": {
                            "listChanged": True
                        },
                        "sampling": {}
                    },
                    "clientInfo": {
                        "name": "brewsource-test-client",
                        "version": "1.0.0"
                    }
                }
            }

            await websocket.send(json.dumps(init_request))
            response = await websocket.recv()
            init_result = json.loads(response)
            print("✅ Initialize:", "✓" if init_result.get("result") else "✗")

            # Test 2: List available tools
            tools_request = {
                "jsonrpc": "2.0",
                "id": 2,
                "method": "tools/list",
                "params": {}
            }

            await websocket.send(json.dumps(tools_request))
            response = await websocket.recv()
            tools_result = json.loads(response)

            if tools_result.get("result") and "tools" in tools_result["result"]:
                tools = tools_result["result"]["tools"]
                print(f"✅ Available tools ({len(tools)}):")
                for tool in tools:
                    print(f"   - {tool['name']}: {tool.get('description', 'No description')}")
            else:
                print("✗ Failed to get tools list")
                return

            print("\n" + "=" * 50)
            print("🧪 Testing Phase 1 Tools")
            print("=" * 50)

            # Test 3: BJCP Lookup
            print("\n1. Testing bjcp_lookup...")
            bjcp_request = {
                "jsonrpc": "2.0",
                "id": 3,
                "method": "tools/call",
                "params": {
                    "name": "bjcp_lookup",
                    "arguments": {
                        "style_code": "21A"
                    }
                }
            }

            await websocket.send(json.dumps(bjcp_request))
            response = await websocket.recv()
            bjcp_result = json.loads(response)

            if bjcp_result.get("result"):
                print("   ✅ BJCP Lookup successful")
                content = bjcp_result["result"]["content"][0]["text"]
                # Extract style name from content
                if "American IPA" in content or "21A" in content:
                    print("   📋 Found American IPA (21A) details")
                else:
                    print(f"   📋 Result: {content[:100]}...")
            else:
                print("   ✗ BJCP Lookup failed")
                print(f"   Error: {bjcp_result.get('error', 'Unknown error')}")

            # Test 4: Search Beers
            print("\n2. Testing search_beers...")
            beer_request = {
                "jsonrpc": "2.0",
                "id": 4,
                "method": "tools/call",
                "params": {
                    "name": "search_beers",
                    "arguments": {
                        "query": "IPA"
                    }
                }
            }

            await websocket.send(json.dumps(beer_request))
            response = await websocket.recv()
            beer_result = json.loads(response)

            if beer_result.get("result"):
                print("   ✅ Beer search successful")
                content = beer_result["result"]["content"][0]["text"]
                if "IPA" in content or "beer" in content.lower():
                    print("   🍺 Found beer results")
                else:
                    print(f"   🍺 Result: {content[:100]}...")
            else:
                print("   ✗ Beer search failed")
                print(f"   Error: {beer_result.get('error', 'Unknown error')}")

            # Test 5: Find Breweries
            print("\n3. Testing find_breweries...")
            brewery_request = {
                "jsonrpc": "2.0",
                "id": 5,
                "method": "tools/call",
                "params": {
                    "name": "find_breweries",
                    "arguments": {
                        "location": "California"
                    }
                }
            }

            await websocket.send(json.dumps(brewery_request))
            response = await websocket.recv()
            brewery_result = json.loads(response)

            if brewery_result.get("result"):
                print("   ✅ Brewery search successful")
                content = brewery_result["result"]["content"][0]["text"]
                if "brewery" in content.lower() or "california" in content.lower():
                    print("   🏭 Found brewery results")
                else:
                    print(f"   🏭 Result: {content[:100]}...")
            else:
                print("   ✗ Brewery search failed")
                print(f"   Error: {brewery_result.get('error', 'Unknown error')}")

            print("\n" + "=" * 50)
            print("🎉 Phase 1 Testing Complete!")
            print("=" * 50)

    except Exception as e:
        print(f"❌ Connection failed: {e}")
        print("Make sure the MCP server is running on ws://localhost:8080/mcp")

if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "--install-deps":
        import subprocess
        subprocess.run([sys.executable, "-m", "pip", "install", "websockets"])
        print("Dependencies installed. Run again without --install-deps")
    else:
        asyncio.run(test_mcp_server())
