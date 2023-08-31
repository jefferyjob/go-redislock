# Version update record

## v1.0.0
- Use Redis backend storage to ensure the stability and reliability of distributed locks
- Provides an easy-to-use API to easily implement functions such as lock, unlock, spin lock, automatic renewal and manual renewal
- Support custom timeout and automatic renewal, flexible configuration according to actual needs

## v1.0.1
- Optimized Lua script for renewal (#22)
