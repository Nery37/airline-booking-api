-- Seed data for airline booking system
-- This file contains sample data for testing and development

-- Insert sample flights
INSERT INTO flights (id, origin, destination, departure_time, arrival_time, airline, aircraft, fare_class) VALUES
(1, 'JFK', 'LAX', '2025-08-30 10:30:00', '2025-08-30 14:30:00', 'AA', 'Boeing 737', 'economy'),
(2, 'JFK', 'LAX', '2025-08-30 15:45:00', '2025-08-30 19:45:00', 'DL', 'Airbus A320', 'economy'),
(3, 'LAX', 'JFK', '2025-08-30 08:15:00', '2025-08-30 16:45:00', 'AA', 'Boeing 777', 'business'),
(4, 'LAX', 'JFK', '2025-08-30 22:00:00', '2025-08-31 06:30:00', 'UA', 'Boeing 787', 'economy'),
(5, 'MIA', 'LAS', '2025-08-30 11:20:00', '2025-08-30 14:10:00', 'SW', 'Boeing 737', 'economy'),
(6, 'LAS', 'MIA', '2025-08-30 16:30:00', '2025-08-30 23:45:00', 'SW', 'Boeing 737', 'economy'),
(7, 'ORD', 'SFO', '2025-08-30 09:00:00', '2025-08-30 11:30:00', 'UA', 'Boeing 757', 'economy'),
(8, 'SFO', 'ORD', '2025-08-30 13:15:00', '2025-08-30 19:20:00', 'AA', 'Airbus A321', 'business'),
(9, 'ATL', 'SEA', '2025-08-30 07:45:00', '2025-08-30 10:15:00', 'DL', 'Boeing 767', 'economy'),
(10, 'SEA', 'ATL', '2025-08-30 18:30:00', '2025-08-31 02:10:00', 'DL', 'Boeing 767', 'first'),
(11, 'JFK', 'LHR', '2025-08-30 21:00:00', '2025-08-31 08:30:00', 'BA', 'Boeing 777', 'business'),
(12, 'LHR', 'JFK', '2025-08-30 14:20:00', '2025-08-30 17:45:00', 'VS', 'Airbus A350', 'economy'),
(13, 'LAX', 'NRT', '2025-08-30 11:45:00', '2025-08-31 16:20:00', 'JL', 'Boeing 787', 'business'),
(14, 'NRT', 'LAX', '2025-08-30 17:30:00', '2025-08-30 10:15:00', 'ANA', 'Boeing 777', 'economy'),
(15, 'MIA', 'GRU', '2025-08-30 23:55:00', '2025-08-31 09:40:00', 'AA', 'Boeing 777', 'economy');

-- Insert sample seats for each flight
-- Flight 1: JFK-LAX (Boeing 737) - 180 seats
INSERT INTO seats (flight_id, seat_no, class) VALUES
-- First Class (rows 1-3, 4 seats per row)
(1, '1A', 'first'), (1, '1B', 'first'), (1, '1C', 'first'), (1, '1D', 'first'),
(1, '2A', 'first'), (1, '2B', 'first'), (1, '2C', 'first'), (1, '2D', 'first'),
(1, '3A', 'first'), (1, '3B', 'first'), (1, '3C', 'first'), (1, '3D', 'first'),
-- Business Class (rows 4-8, 4 seats per row)
(1, '4A', 'business'), (1, '4B', 'business'), (1, '4C', 'business'), (1, '4D', 'business'),
(1, '5A', 'business'), (1, '5B', 'business'), (1, '5C', 'business'), (1, '5D', 'business'),
(1, '6A', 'business'), (1, '6B', 'business'), (1, '6C', 'business'), (1, '6D', 'business'),
(1, '7A', 'business'), (1, '7B', 'business'), (1, '7C', 'business'), (1, '7D', 'business'),
(1, '8A', 'business'), (1, '8B', 'business'), (1, '8C', 'business'), (1, '8D', 'business'),
-- Economy Class (rows 9-35, 6 seats per row)
(1, '9A', 'economy'), (1, '9B', 'economy'), (1, '9C', 'economy'), (1, '9D', 'economy'), (1, '9E', 'economy'), (1, '9F', 'economy'),
(1, '10A', 'economy'), (1, '10B', 'economy'), (1, '10C', 'economy'), (1, '10D', 'economy'), (1, '10E', 'economy'), (1, '10F', 'economy'),
(1, '11A', 'economy'), (1, '11B', 'economy'), (1, '11C', 'economy'), (1, '11D', 'economy'), (1, '11E', 'economy'), (1, '11F', 'economy'),
(1, '12A', 'economy'), (1, '12B', 'economy'), (1, '12C', 'economy'), (1, '12D', 'economy'), (1, '12E', 'economy'), (1, '12F', 'economy'),
(1, '13A', 'economy'), (1, '13B', 'economy'), (1, '13C', 'economy'), (1, '13D', 'economy'), (1, '13E', 'economy'), (1, '13F', 'economy'),
(1, '14A', 'economy'), (1, '14B', 'economy'), (1, '14C', 'economy'), (1, '14D', 'economy'), (1, '14E', 'economy'), (1, '14F', 'economy'),
(1, '15A', 'economy'), (1, '15B', 'economy'), (1, '15C', 'economy'), (1, '15D', 'economy'), (1, '15E', 'economy'), (1, '15F', 'economy'),
(1, '16A', 'economy'), (1, '16B', 'economy'), (1, '16C', 'economy'), (1, '16D', 'economy'), (1, '16E', 'economy'), (1, '16F', 'economy'),
(1, '17A', 'economy'), (1, '17B', 'economy'), (1, '17C', 'economy'), (1, '17D', 'economy'), (1, '17E', 'economy'), (1, '17F', 'economy'),
(1, '18A', 'economy'), (1, '18B', 'economy'), (1, '18C', 'economy'), (1, '18D', 'economy'), (1, '18E', 'economy'), (1, '18F', 'economy'),
(1, '19A', 'economy'), (1, '19B', 'economy'), (1, '19C', 'economy'), (1, '19D', 'economy'), (1, '19E', 'economy'), (1, '19F', 'economy'),
(1, '20A', 'economy'), (1, '20B', 'economy'), (1, '20C', 'economy'), (1, '20D', 'economy'), (1, '20E', 'economy'), (1, '20F', 'economy'),
(1, '21A', 'economy'), (1, '21B', 'economy'), (1, '21C', 'economy'), (1, '21D', 'economy'), (1, '21E', 'economy'), (1, '21F', 'economy'),
(1, '22A', 'economy'), (1, '22B', 'economy'), (1, '22C', 'economy'), (1, '22D', 'economy'), (1, '22E', 'economy'), (1, '22F', 'economy'),
(1, '23A', 'economy'), (1, '23B', 'economy'), (1, '23C', 'economy'), (1, '23D', 'economy'), (1, '23E', 'economy'), (1, '23F', 'economy'),
(1, '24A', 'economy'), (1, '24B', 'economy'), (1, '24C', 'economy'), (1, '24D', 'economy'), (1, '24E', 'economy'), (1, '24F', 'economy'),
(1, '25A', 'economy'), (1, '25B', 'economy'), (1, '25C', 'economy'), (1, '25D', 'economy'), (1, '25E', 'economy'), (1, '25F', 'economy'),
(1, '26A', 'economy'), (1, '26B', 'economy'), (1, '26C', 'economy'), (1, '26D', 'economy'), (1, '26E', 'economy'), (1, '26F', 'economy'),
(1, '27A', 'economy'), (1, '27B', 'economy'), (1, '27C', 'economy'), (1, '27D', 'economy'), (1, '27E', 'economy'), (1, '27F', 'economy'),
(1, '28A', 'economy'), (1, '28B', 'economy'), (1, '28C', 'economy'), (1, '28D', 'economy'), (1, '28E', 'economy'), (1, '28F', 'economy'),
(1, '29A', 'economy'), (1, '29B', 'economy'), (1, '29C', 'economy'), (1, '29D', 'economy'), (1, '29E', 'economy'), (1, '29F', 'economy'),
(1, '30A', 'economy'), (1, '30B', 'economy'), (1, '30C', 'economy'), (1, '30D', 'economy'), (1, '30E', 'economy'), (1, '30F', 'economy');

-- Flight 2: JFK-LAX (Airbus A320) - 150 seats
INSERT INTO seats (flight_id, seat_no, class) VALUES
-- Business Class (rows 1-4, 4 seats per row)
(2, '1A', 'business'), (2, '1B', 'business'), (2, '1C', 'business'), (2, '1D', 'business'),
(2, '2A', 'business'), (2, '2B', 'business'), (2, '2C', 'business'), (2, '2D', 'business'),
(2, '3A', 'business'), (2, '3B', 'business'), (2, '3C', 'business'), (2, '3D', 'business'),
(2, '4A', 'business'), (2, '4B', 'business'), (2, '4C', 'business'), (2, '4D', 'business'),
-- Economy Class (rows 5-29, 6 seats per row)
(2, '5A', 'economy'), (2, '5B', 'economy'), (2, '5C', 'economy'), (2, '5D', 'economy'), (2, '5E', 'economy'), (2, '5F', 'economy'),
(2, '6A', 'economy'), (2, '6B', 'economy'), (2, '6C', 'economy'), (2, '6D', 'economy'), (2, '6E', 'economy'), (2, '6F', 'economy'),
(2, '7A', 'economy'), (2, '7B', 'economy'), (2, '7C', 'economy'), (2, '7D', 'economy'), (2, '7E', 'economy'), (2, '7F', 'economy'),
(2, '8A', 'economy'), (2, '8B', 'economy'), (2, '8C', 'economy'), (2, '8D', 'economy'), (2, '8E', 'economy'), (2, '8F', 'economy'),
(2, '9A', 'economy'), (2, '9B', 'economy'), (2, '9C', 'economy'), (2, '9D', 'economy'), (2, '9E', 'economy'), (2, '9F', 'economy'),
(2, '10A', 'economy'), (2, '10B', 'economy'), (2, '10C', 'economy'), (2, '10D', 'economy'), (2, '10E', 'economy'), (2, '10F', 'economy'),
(2, '11A', 'economy'), (2, '11B', 'economy'), (2, '11C', 'economy'), (2, '11D', 'economy'), (2, '11E', 'economy'), (2, '11F', 'economy'),
(2, '12A', 'economy'), (2, '12B', 'economy'), (2, '12C', 'economy'), (2, '12D', 'economy'), (2, '12E', 'economy'), (2, '12F', 'economy'),
(2, '13A', 'economy'), (2, '13B', 'economy'), (2, '13C', 'economy'), (2, '13D', 'economy'), (2, '13E', 'economy'), (2, '13F', 'economy'),
(2, '14A', 'economy'), (2, '14B', 'economy'), (2, '14C', 'economy'), (2, '14D', 'economy'), (2, '14E', 'economy'), (2, '14F', 'economy'),
(2, '15A', 'economy'), (2, '15B', 'economy'), (2, '15C', 'economy'), (2, '15D', 'economy'), (2, '15E', 'economy'), (2, '15F', 'economy'),
(2, '16A', 'economy'), (2, '16B', 'economy'), (2, '16C', 'economy'), (2, '16D', 'economy'), (2, '16E', 'economy'), (2, '16F', 'economy'),
(2, '17A', 'economy'), (2, '17B', 'economy'), (2, '17C', 'economy'), (2, '17D', 'economy'), (2, '17E', 'economy'), (2, '17F', 'economy'),
(2, '18A', 'economy'), (2, '18B', 'economy'), (2, '18C', 'economy'), (2, '18D', 'economy'), (2, '18E', 'economy'), (2, '18F', 'economy'),
(2, '19A', 'economy'), (2, '19B', 'economy'), (2, '19C', 'economy'), (2, '19D', 'economy'), (2, '19E', 'economy'), (2, '19F', 'economy'),
(2, '20A', 'economy'), (2, '20B', 'economy'), (2, '20C', 'economy'), (2, '20D', 'economy'), (2, '20E', 'economy'), (2, '20F', 'economy'),
(2, '21A', 'economy'), (2, '21B', 'economy'), (2, '21C', 'economy'), (2, '21D', 'economy'), (2, '21E', 'economy'), (2, '21F', 'economy'),
(2, '22A', 'economy'), (2, '22B', 'economy'), (2, '22C', 'economy'), (2, '22D', 'economy'), (2, '22E', 'economy'), (2, '22F', 'economy'),
(2, '23A', 'economy'), (2, '23B', 'economy'), (2, '23C', 'economy'), (2, '23D', 'economy'), (2, '23E', 'economy'), (2, '23F', 'economy'),
(2, '24A', 'economy'), (2, '24B', 'economy'), (2, '24C', 'economy'), (2, '24D', 'economy'), (2, '24E', 'economy'), (2, '24F', 'economy'),
(2, '25A', 'economy'), (2, '25B', 'economy'), (2, '25C', 'economy'), (2, '25D', 'economy'), (2, '25E', 'economy'), (2, '25F', 'economy');

-- Simplified seat data for other flights (just a few seats per flight for demo)
-- Flight 3: LAX-JFK (Boeing 777 Business)
INSERT INTO seats (flight_id, seat_no, class) VALUES
(3, '1A', 'business'), (3, '1B', 'business'), (3, '1C', 'business'), (3, '1D', 'business'),
(3, '2A', 'business'), (3, '2B', 'business'), (3, '2C', 'business'), (3, '2D', 'business'),
(3, '3A', 'business'), (3, '3B', 'business'), (3, '3C', 'business'), (3, '3D', 'business'),
(3, '10A', 'economy'), (3, '10B', 'economy'), (3, '10C', 'economy'), (3, '10D', 'economy'), (3, '10E', 'economy'), (3, '10F', 'economy'),
(3, '11A', 'economy'), (3, '11B', 'economy'), (3, '11C', 'economy'), (3, '11D', 'economy'), (3, '11E', 'economy'), (3, '11F', 'economy'),
(3, '12A', 'economy'), (3, '12B', 'economy'), (3, '12C', 'economy'), (3, '12D', 'economy'), (3, '12E', 'economy'), (3, '12F', 'economy');

-- Flight 4: LAX-JFK (Boeing 787 Economy)
INSERT INTO seats (flight_id, seat_no, class) VALUES
(4, '15A', 'economy'), (4, '15B', 'economy'), (4, '15C', 'economy'), (4, '15D', 'economy'), (4, '15E', 'economy'), (4, '15F', 'economy'),
(4, '16A', 'economy'), (4, '16B', 'economy'), (4, '16C', 'economy'), (4, '16D', 'economy'), (4, '16E', 'economy'), (4, '16F', 'economy'),
(4, '17A', 'economy'), (4, '17B', 'economy'), (4, '17C', 'economy'), (4, '17D', 'economy'), (4, '17E', 'economy'), (4, '17F', 'economy'),
(4, '18A', 'economy'), (4, '18B', 'economy'), (4, '18C', 'economy'), (4, '18D', 'economy'), (4, '18E', 'economy'), (4, '18F', 'economy');

-- Flight 5: MIA-LAS (Southwest Boeing 737)
INSERT INTO seats (flight_id, seat_no, class) VALUES
(5, '5A', 'economy'), (5, '5B', 'economy'), (5, '5C', 'economy'), (5, '5D', 'economy'), (5, '5E', 'economy'), (5, '5F', 'economy'),
(5, '6A', 'economy'), (5, '6B', 'economy'), (5, '6C', 'economy'), (5, '6D', 'economy'), (5, '6E', 'economy'), (5, '6F', 'economy'),
(5, '7A', 'economy'), (5, '7B', 'economy'), (5, '7C', 'economy'), (5, '7D', 'economy'), (5, '7E', 'economy'), (5, '7F', 'economy');

-- Add some sample holds (seats that are temporarily held)
INSERT INTO seat_locks (flight_id, seat_no, holder_id, expires_at) VALUES
(1, '12A', 'user123', DATE_ADD(NOW(), INTERVAL 10 MINUTE)),
(1, '15C', 'user456', DATE_ADD(NOW(), INTERVAL 5 MINUTE)),
(2, '8B', 'user789', DATE_ADD(NOW(), INTERVAL 12 MINUTE)),
(3, '2A', 'user111', DATE_ADD(NOW(), INTERVAL 8 MINUTE));

-- Add some sample confirmed tickets
INSERT INTO tickets (flight_id, seat_no, user_id, price_amount, currency, pnr_code, payment_ref) VALUES
(1, '10A', 'customer001', 29900, 'USD', 'ABC001', 'pay_001_12345'),
(1, '10B', 'customer002', 29900, 'USD', 'ABC002', 'pay_002_12346'),
(1, '1A', 'customer003', 149900, 'USD', 'ABC003', 'pay_003_12347'),
(2, '5A', 'customer004', 31900, 'USD', 'DEF001', 'pay_004_12348'),
(2, '5B', 'customer005', 31900, 'USD', 'DEF002', 'pay_005_12349'),
(3, '1A', 'customer006', 89900, 'USD', 'GHI001', 'pay_006_12350'),
(3, '10A', 'customer007', 49900, 'USD', 'GHI002', 'pay_007_12351'),
(4, '15A', 'customer008', 27900, 'USD', 'JKL001', 'pay_008_12352'),
(5, '5A', 'customer009', 19900, 'USD', 'MNO001', 'pay_009_12353');
