class WyF3 {
    static DefaultSeed = BigInt.asUintN(64, 0xa0761d6478bd642fn);
    static s1 = BigInt.asUintN(64, 0xe7037ed1a0b428dbn);
    static s2 = BigInt.asUintN(64, 0x8ebc6af09c88c6e3n);
    static s3 = BigInt.asUintN(64, 0x589965cc75374cc3n);
    static s4 = BigInt.asUintN(64, 0x1d8e4e27c47d124fn);
    static mask64 = BigInt.asUintN(64, 0xFFFFFFFFFFFFFFFFn);
    static mask128 = BigInt.asUintN(128, 0xFFFFFFFFFFFFFFFFn);
    static uint64hi = BigInt.asUintN(64, WyF3.s1 ^ BigInt.asUintN(64, 8n));
    static numberBuffer = Uint8Array.of(0, 0, 0, 0, 0, 0, 0, 0);
    static numberBufferView = new DataView(WyF3.numberBuffer.buffer);

    static mix(x, y) {
        // mul128
        let r = BigInt.asUintN(128, x) * BigInt.asUintN(128, y);
        let hi = BigInt.asUintN(64, (r >> 64n) & WyF3.mask128);
        let lo = BigInt.asUintN(64, r & WyF3.mask128);
        return hi ^ lo;
    }

    static read32(d, offset) {
        return BigInt.asUintN(64, BigInt(d.getUint32(offset, true)));
    }

    static uint64(value = 0n, seed = WyF3.DefaultSeed) {
        WyF3.numberBufferView.setBigUint64(0, BigInt.asUintN(64, BigInt(value)), true);
        return WyF3.uint64Internal(seed);
    }

    static number(value = 0, seed = WyF3.DefaultSeed) {
        WyF3.numberBufferView.setFloat64(0, value, true);
        return WyF3.uint64Internal(seed);
    }

    static uint64Internal(seed = WyF3.DefaultSeed) {
        const v = WyF3.numberBufferView.getBigUint64(0, true);
        let hi = BigInt.asUintN(64, (v >> 32n) & WyF3.mask64);
        let lo = BigInt.asUintN(64, v & WyF3.mask64);
        return WyF3.mix(WyF3.uint64hi, WyF3.mix(hi ^ WyF3.s1, lo ^ seed));
    }

    static string(data = '', seed = WyF3.DefaultSeed) {
        const d = new TextEncoder().encode(data);
        return WyF3.hashBytes(new DataView(d.buffer), seed);
    }

    static hash(value, seed = WyF3.DefaultSeed) {
        if (typeof value === 'string') {
            return WyF3.string(value, seed);
        }
        if (typeof value === 'bigint') {
            return WyF3.uint64(value, seed);
        }
        if (typeof value === 'number') {
            return WyF3.number(value, seed);
        }
        if (value instanceof DataView) {
            return WyF3.hashBytes(value, seed);
        }
        if (value instanceof ArrayBuffer) {
            return WyF3.hashBytes(new DataView(value), seed);
        }
        if (value instanceof Uint8Array) {
            return WyF3.hashBytes(new DataView(value.buffer), seed);
        }
        throw 'WyF3.hash supports string, bigint, number, DataView, ArrayBuffer, Uint8Array';
    }

    /**
     *
     * @param d DataView
     * @param seed
     * @returns {*}
     */
    static hashBytes(d, seed = WyF3.DefaultSeed) {
        let length = d.byteLength;
        let a;
        let b;

        if (length <= 16) {
            if (length >= 4) {
                a = (WyF3.read32(d, 0) << BigInt.asUintN(64, 32n)) | WyF3.read32(d, (length >> 3) << 2);
                b = (WyF3.read32(d, length - 4) << BigInt.asUintN(64, 32n)) | WyF3.read32(d, length - 4 - ((length >> 3) << 2));
            } else if (length > 0) {
                a = BigInt.asUintN(64, BigInt(d.getUint8(0) << 16 |
                    d.getUint8(length >> 1) << 8 |
                    d.getUint8(length - 1)));
                b = 0n;
            }
        } else {
            let index = length;
            let start = 0;
            if (length > 48) {
                let see1 = seed;
                let see2 = seed;
                for (; index > 48;) {
                    seed = WyF3.mix(d.getBigUint64(start, true) ^ WyF3.s1, d.getBigUint64(start + 8, true) ^ seed);
                    see1 = WyF3.mix(d.getBigUint64(start + 16, true) ^ WyF3.s2, d.getBigUint64(start + 24, true) ^ see1);
                    see2 = WyF3.mix(d.getBigUint64(start + 32, true) ^ WyF3.s3, d.getBigUint64(start + 40, true) ^ see2);
                    index -= 48;
                    start += 48;
                }
                seed ^= see1 ^ see2;
            }

            for (; index > 16;) {
                seed = WyF3.mix(d.getBigUint64(start, true) ^ WyF3.s1, d.getBigUint64(start + 8, true) ^ seed);
                index -= 16;
                start += 16;
            }

            a = d.getBigUint64(start + index - 16, true);
            b = d.getBigUint64(start + index - 8, true);
        }

        return WyF3.mix(WyF3.s1 ^ BigInt(length), WyF3.mix(a ^ WyF3.s1, b ^ seed));
    }
}

class WyF3Rand {
    static DEFAULT = new WyF3Rand(new Date().getTime());

    constructor(seed) {
        this.seed = BigInt.asUintN(64, BigInt(seed));
    }

    static RAND_C0 = BigInt.asUintN(64, 0xa0761d6478bd642fn);
    static RAND_C1 = BigInt.asUintN(64, 0xe7037ed1a0b428dbn);


    next() {
        this.seed += WyF3Rand.RAND_C0;
        return WyF3.mix(this.seed, this.seed ^ WyF3Rand.RAND_C1);
    }

    // static NORM_01 = 1.0 / (1.0 << 52);
    // static GAUSSIAN_NORM = 1.0 / (1.0 << 20);
    //
    // next01() {
    //     WyF3.numberBufferView.setBigUint64(0, this.next(), true);
    //     let r = WyF3.numberBufferView.getFloat64(0, true);
    //     return (r >> 12) * WyF3Rand.NORM_01;
    // }
    //
    // nextGaussian() {
    //     WyF3.numberBufferView.setBigUint64(0, this.next(), true);
    //     let r = WyF3.numberBufferView.getBigUint64(0, true);
    //     WyF3.numberBufferView.setBigUint64(0, (r & 0x1fffffn) + ((r >> 21n) & 0x1fffffn) + ((r >> 42n) & 0x1fffffn), true);
    //     return WyF3.numberBufferView.getFloat64(0, true) * WyF3Rand.GAUSSIAN_NORM
    //     // return ((r & 0x1fffff) + ((r >> 21) & 0x1fffff) + ((r >> 42) & 0x1fffff)) * WyF3Rand.GAUSSIAN_NORM - 3.0;
    // }
}
