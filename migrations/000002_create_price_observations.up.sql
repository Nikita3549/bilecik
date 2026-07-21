CREATE TABLE price_observations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_iata CHAR(3) NOT NULL,
    to_iata CHAR(3) NOT NULL,
    flight_date DATE NOT NULL,
    amount NUMERIC(10, 2) NOT NULL,
    currency CHAR(3) NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL,
    checked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_observations_route_date ON price_observations (
    from_iata, to_iata, flight_date
);
CREATE INDEX idx_observations_checked ON price_observations (
    observed_at
) WHERE checked
= FALSE;
