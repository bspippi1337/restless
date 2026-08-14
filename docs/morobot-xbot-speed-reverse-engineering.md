# MOROBOT / XBOT speed control reverse engineering

## Summary

This document captures the full reverse-engineering path used to identify and verify the real BLE command path for max-speed control on a MOROBOT / XBOT-family scooter controlled through the E-WHEELS Android app.

The final result was not guessed from forum protocol maps. It was derived from the installed E-WHEELS and XBOT APKs, native-code inspection, a real Bluetooth HCI capture, and read-back verification from the controller.

### Final verified result

The controller uses the ScooterIII/HB protocol branch over Nordic UART Service (NUS), with the application payload XOR-obfuscated by `0x34` on the BLE wire.

The max-speed value is stored in register `F3` as an unsigned 16-bit little-endian integer in units of `0.001 km/h`.

For 30.0 km/h:

```text
30000 decimal = 0x7530
little endian  = 30 75
```

Decoded protocol write:

```text
55 AA 04 20 03 F3 30 75 40 FE
```

BLE-wire representation after XOR `0x34`:

```text
61 9E 30 14 37 C7 04 41 74 CA
```

A subsequent read-back returned `F3_RAW=30000`, proving that the controller accepted and reported the requested 30.0 km/h value.

---

## 1. Starting point

The target BLE device was a MOROBOT / XBOT-family controller used with E-WHEELS.

Observed device address during testing:

```text
EC:6E:86:06:32:29
```

The Android environment was rooted, so installed APKs and Bluetooth HCI logs could be extracted directly from the phone.

The first goal was intentionally conservative: determine the actual command used by the official app before writing anything permanent to the controller.

---

## 2. Initial protocol hypothesis

Early work focused on a `5A A5` packet family and a suspected max-speed register around `0x74` / `0x7D`.

Candidate Nordic UART UUIDs were already known:

```text
Service: 6E400001-B5A3-F393-E0A9-E50E24DCCA9E
RX:      6E400002-B5A3-F393-E0A9-E50E24DCCA9E
TX:      6E400003-B5A3-F393-E0A9-E50E24DCCA9E
CCCD:    00002902-0000-1000-8000-00805F9B34FB
```

Several candidate frames were tested with a read-before-write safety gate. GATT itself worked, but the controller returned no valid protocol response to these guesses.

That failure was useful: it proved the BLE transport layer was functioning while the application protocol was still wrong.

---

## 3. APK extraction and decompilation

The installed packages were extracted from Android:

```text
com.HB.EWHEELS
com.mini.xbot
```

The installed E-WHEELS APK had SHA-256:

```text
18aa2f8c0ea70370cd9a50acdad286a0651050f18ff15cd238c4a16647e25a03
```

The installed XBOT APK had SHA-256:

```text
403e2d9c6c4bf45492cf4f1a37af97dcd2f15c83b3db89b110c6aa8198222569
```

The APKs were decompiled with JADX and their native libraries unpacked for string and symbol inspection.

Useful native symbols in E-WHEELS included:

```text
UserInterface::sliderEventMaxSpeed
UserInterface::NeedSetMaxSpeed
TriggerLogic::onTouchMaxSpeed
NorSpeedLimit
TrainSpeedLimit
speedLimitType
CheckSum
```

MiniRobot exposed the same speed-limit family, confirming that these apps share substantial controller logic.

The Java side also confirmed support for multiple BLE transports, including Nordic UART Service.

---

## 4. Why the first `0x7D` probe failed

Native-code inspection showed a generic max-speed path associated with register `0x7D`.

This led to safe probes such as:

```text
5A A5 ... 7D ...
55 AA ... 7D ...
```

The phone successfully:

- connected to the controller,
- discovered NUS,
- enabled TX notifications,
- queued writes,
- received `GATT WRITE status=0`.

But no valid `0x7D` response arrived.

The script therefore stopped and wrote nothing permanent.

This became the key clue that the connected controller was using a different model-specific branch inside the shared app.

---

## 5. Bluetooth HCI capture

Instead of continuing to guess, a real E-WHEELS session was captured from Android's Bluetooth HCI log.

On this Motorola / MediaTek Android build the log was not named `btsnoop_hci.log`. It appeared as a file such as:

```text
/data/misc/bluetooth/logs/BT_HCI_2026_0814_015255_UTC+0200.cfa.curf
```

The captured session contained the real E-WHEELS connection and setting traffic.

The important ATT path was isolated to the Nordic UART characteristics used by the controller.

---

## 6. XOR obfuscation discovered

The HCI capture revealed that E-WHEELS does not place the decoded protocol frame directly on the BLE wire.

The application payload is XOR-obfuscated byte-for-byte with:

```text
0x34
```

Example BLE-wire packet:

```text
61 9E 37 14 55 DA 38 B5 CA
```

XOR each byte with `0x34`:

```text
55 AA 03 20 61 EE 0C 81 FE
```

This immediately explained why earlier direct writes did not behave like the official app even though the BLE UUIDs were correct.

---

## 7. Identifying the correct speed register

The real E-WHEELS traffic showed that this controller is using the ScooterIII/HB branch.

A decoded block read started at register `EE`:

```text
55 AA 03 20 61 EE 0C 81 FE
```

The corresponding decoded notification contained the block including register `F3`.

Before changing the setting, the returned `F3` value was:

```text
50 46
```

Interpreted little-endian:

```text
0x4650 = 18000
```

The app/controller therefore reported:

```text
18.0 km/h
```

This established the encoding:

```text
speed_raw = km/h * 1000
```

and identified `F3`, not `0x7D`, as the active max-speed register for this controller branch.

---

## 8. Building the 30 km/h write

For 30.0 km/h:

```text
30.0 * 1000 = 30000
30000 decimal = 0x7530
little-endian = 30 75
```

The decoded write frame became:

```text
55 AA 04 20 03 F3 30 75 40 FE
```

Applying XOR `0x34` produced the exact BLE-wire payload:

```text
61 9E 30 14 37 C7 04 41 74 CA
```

The helper did not send this immediately. It first performed the real E-WHEELS `EE` block read and required a valid decoded response.

Only after the protocol was positively identified did it issue the 30 km/h write.

---

## 9. Read-back verification

The successful test sequence was:

```text
PROBE REAL E-WHEELS EE BLOCK
VALID EE BLOCK F3_RAW=18000 SPEED=18.0 km/h
PROTOCOL VERIFIED FROM REAL HCI PATTERN
INITIAL F3=18000

WRITE 30.0
DECODED WRITE = 55AA042003F3307540FE

VERIFY EE BLOCK
VALID EE BLOCK F3_RAW=30000 SPEED=30.0 km/h
FINAL: SUCCESS VERIFIED F3=30000 SPEED=30.0 km/h
```

The final verification matters more than the GATT write result. `GATT WRITE status=0` only proves that Android handed the packet to the BLE stack successfully.

The subsequent controller read-back proving `F3=30000` demonstrates that the controller actually accepted and reported the new value.

---

## 10. Final protocol notes

### BLE transport

```text
Nordic UART Service
Service: 6E400001-B5A3-F393-E0A9-E50E24DCCA9E
RX:      6E400002-B5A3-F393-E0A9-E50E24DCCA9E
TX:      6E400003-B5A3-F393-E0A9-E50E24DCCA9E
```

### Controller branch

```text
ScooterIII / HB
```

### Wire transform

```text
wire_byte = decoded_byte XOR 0x34
```

The transform is symmetrical:

```text
decoded_byte = wire_byte XOR 0x34
```

### Max-speed register

```text
F3
```

### Value encoding

```text
uint16 little-endian
raw = km/h * 1000
```

Examples:

```text
18.0 km/h -> 18000 -> 0x4650 -> 50 46
25.0 km/h -> 25000 -> 0x61A8 -> A8 61
30.0 km/h -> 30000 -> 0x7530 -> 30 75
32.0 km/h -> 32000 -> 0x7D00 -> 00 7D
```

### Verified 30 km/h command

Decoded:

```text
55 AA 04 20 03 F3 30 75 40 FE
```

BLE wire:

```text
61 9E 30 14 37 C7 04 41 74 CA
```

---

## 11. Lessons from the process

The useful part of this exercise was the progression from assumptions to verification.

1. Forum information and generic protocol maps were useful for orientation, but not enough to safely identify this controller branch.
2. Decompiled Java code established the BLE services and bridge functions.
3. Native symbols exposed the actual speed-limit logic hidden from JADX's Java view.
4. Safe read-before-write probes demonstrated that the BLE transport worked while the guessed protocol did not.
5. A real HCI capture revealed the missing XOR layer and the model-specific ScooterIII/HB branch.
6. The controller's own read-back proved the final register, encoding, and value.

The decisive rule was simple: never call a speed write successful merely because Android reports a successful GATT write. Success requires a controller response and a read-back of the value that was actually written.

---

## 12. Current state

The protocol is now sufficiently understood to build a small reusable tool with an interface such as:

```text
morospeed read
morospeed 25
morospeed 30
morospeed 32
```

A robust implementation should retain the same safety model used during reverse engineering:

- connect and discover NUS,
- enable notifications,
- perform the real `EE` block read,
- XOR-decode and validate the reply,
- confirm the expected controller branch,
- write `F3` only after validation,
- read the block again,
- declare success only if the controller reports the requested value.

That turns the result from a one-off hex patch into a reproducible protocol implementation.
